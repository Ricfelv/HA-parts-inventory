package web

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ha-parts-inventory/internal/parts"
)

//go:embed templates/*.html static/style.css static/app.js
var assets embed.FS

type Handler struct {
	service   *parts.Service
	templates *template.Template
	css       []byte
	js        []byte
}

type inventoryPage struct {
	Parts       []parts.Part
	BasePath    string
	Query       string
	InStockOnly bool
	Sort        string
	Direction   string
}

type formPage struct {
	Title      string
	Part       parts.Part
	Types      []string
	Error      string
	IsEdit     bool
	FormAction string
	Adjustment string
	BasePath   string
}

func NewHandler(service *parts.Service) (*Handler, error) {
	templates, err := template.New("pages").Funcs(template.FuncMap{"sortURL": sortURL}).ParseFS(assets, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	css, err := fs.ReadFile(assets, "static/style.css")
	if err != nil {
		return nil, fmt.Errorf("read CSS: %w", err)
	}
	js, err := fs.ReadFile(assets, "static/app.js")
	if err != nil {
		return nil, fmt.Errorf("read JavaScript: %w", err)
	}
	return &Handler{service: service, templates: templates, css: css, js: js}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.inventory)
	mux.HandleFunc("GET /parts/new", h.newPart)
	mux.HandleFunc("POST /parts", h.createPart)
	mux.HandleFunc("GET /parts/{id}/edit", h.editPart)
	mux.HandleFunc("POST /parts/{id}", h.updatePart)
	mux.HandleFunc("GET /static/style.css", h.style)
	mux.HandleFunc("GET /static/app.js", h.script)
	return mux
}

func (h *Handler) style(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(h.css)
}

func (h *Handler) script(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write(h.js)
}

func (h *Handler) inventory(w http.ResponseWriter, r *http.Request) {
	options := listOptions(r.URL.Query())
	items, err := h.service.List(r.Context(), options)
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.render(w, http.StatusOK, "inventory", inventoryPage{Parts: items, BasePath: ingressPath(r), Query: options.Query, InStockOnly: options.InStockOnly, Sort: options.Sort, Direction: options.Direction})
}

func (h *Handler) newPart(w http.ResponseWriter, r *http.Request) {
	basePath := ingressPath(r)
	h.renderForm(w, http.StatusOK, formPage{Title: "Add Part", Types: parts.Types, FormAction: basePath + "/parts", BasePath: basePath})
}

func (h *Handler) createPart(w http.ResponseWriter, r *http.Request) {
	input, err := parseInput(r)
	if err == nil {
		_, err = h.service.Create(r.Context(), input)
	}
	if err != nil {
		basePath := ingressPath(r)
		h.renderForm(w, http.StatusBadRequest, formPage{Title: "Add Part", Part: partFromInput(input), Types: parts.Types, Error: displayError(err), FormAction: basePath + "/parts", BasePath: basePath})
		return
	}
	h.redirectInventory(w, r)
}

func (h *Handler) editPart(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	part, err := h.service.Get(r.Context(), id)
	if h.handleGetError(w, err) {
		return
	}
	basePath := ingressPath(r)
	h.renderForm(w, http.StatusOK, formPage{Title: "Edit Part", Part: part, Types: parts.Types, IsEdit: true, FormAction: fmt.Sprintf("%s/parts/%d", basePath, id), BasePath: basePath})
}

func (h *Handler) updatePart(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		h.clientError(w, "Invalid form submission.")
		return
	}
	action := r.FormValue("action")
	var err error
	switch action {
	case "save":
		input, parseErr := inputFromValues(r.Form)
		if parseErr != nil {
			err = parseErr
		} else {
			err = h.service.Update(r.Context(), id, input)
		}
	case "add", "remove":
		amount, parseErr := parseNonNegative(r.FormValue("adjustment"), "Adjustment")
		if parseErr != nil {
			err = parseErr
		} else {
			err = h.service.AdjustQuantity(r.Context(), id, amount, action == "add")
		}
	default:
		err = parts.ValidationError{Message: "Unknown update action."}
	}
	if err != nil {
		if errors.Is(err, parts.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		part, getErr := h.service.Get(r.Context(), id)
		if getErr != nil {
			h.serverError(w, getErr)
			return
		}
		if action == "save" {
			part = partFromForm(r.Form, part)
		}
		basePath := ingressPath(r)
		h.renderForm(w, http.StatusBadRequest, formPage{Title: "Edit Part", Part: part, Types: parts.Types, Error: displayError(err), IsEdit: true, FormAction: fmt.Sprintf("%s/parts/%d", basePath, id), Adjustment: r.FormValue("adjustment"), BasePath: basePath})
		return
	}
	h.redirectInventory(w, r)
}

func listOptions(values url.Values) parts.ListOptions {
	// An absent filter is the initial page state and therefore shows in-stock
	// parts only. The hidden form value explicitly records an unchecked box.
	stock := true
	if selections, present := values["in_stock"]; present {
		stock = false
		for _, selection := range selections {
			if selection == "1" {
				stock = true
				break
			}
		}
	}
	return parts.ListOptions{Query: strings.TrimSpace(values.Get("q")), InStockOnly: stock, Sort: values.Get("sort"), Direction: values.Get("dir")}
}

func parseInput(r *http.Request) (parts.Input, error) {
	if err := r.ParseForm(); err != nil {
		return parts.Input{}, parts.ValidationError{Message: "Invalid form submission."}
	}
	return inputFromValues(r.Form)
}
func inputFromValues(values url.Values) (parts.Input, error) {
	quantity, err := parseNonNegative(values.Get("quantity"), "Quantity")
	if err != nil {
		return parts.Input{}, err
	}
	return parts.Input{Type: values.Get("type"), Value: values.Get("value"), Package: values.Get("package"), Description: values.Get("description"), MPN: values.Get("mpn"), Quantity: quantity}, nil
}
func parseNonNegative(value, field string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, parts.ValidationError{Message: field + " is required."}
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < 0 {
		return 0, parts.ValidationError{Message: field + " must be a whole number zero or greater."}
	}
	return number, nil
}
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return 0, false
	}
	return id, true
}
func partFromInput(input parts.Input) parts.Part {
	return parts.Part{Type: input.Type, Value: input.Value, Package: input.Package, Description: input.Description, MPN: input.MPN, Quantity: input.Quantity}
}
func partFromForm(values url.Values, fallback parts.Part) parts.Part {
	fallback.Type, fallback.Value, fallback.Package, fallback.Description, fallback.MPN = values.Get("type"), values.Get("value"), values.Get("package"), values.Get("description"), values.Get("mpn")
	if quantity, err := strconv.Atoi(values.Get("quantity")); err == nil {
		fallback.Quantity = quantity
	}
	return fallback
}
func (h *Handler) redirectInventory(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, ingressPath(r)+"/", http.StatusSeeOther)
}
func (h *Handler) renderForm(w http.ResponseWriter, status int, page formPage) {
	h.render(w, status, "part_form", page)
}
func (h *Handler) render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}
func (h *Handler) handleGetError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, parts.ErrNotFound) {
		http.NotFound(w, nil)
		return true
	}
	h.serverError(w, err)
	return true
}
func (h *Handler) clientError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}
func (h *Handler) serverError(w http.ResponseWriter, _ error) {
	http.Error(w, "Internal server error.", http.StatusInternalServerError)
}
func displayError(err error) string {
	if validation, ok := err.(parts.ValidationError); ok {
		return validation.Message
	}
	return "Unable to save this part. Please try again."
}

func sortURL(page inventoryPage, field string) string {
	direction := "asc"
	if page.Sort == field && strings.EqualFold(page.Direction, "asc") {
		direction = "desc"
	}
	values := url.Values{"sort": {field}, "dir": {direction}}
	if page.Query != "" {
		values.Set("q", page.Query)
	}
	if page.InStockOnly {
		values.Set("in_stock", "1")
	}
	return page.BasePath + "/?" + values.Encode()
}

// ingressPath supports Home Assistant Ingress and stays empty for standalone use.
func ingressPath(r *http.Request) string {
	path := strings.TrimRight(strings.TrimSpace(r.Header.Get("X-Ingress-Path")), "/")
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return ""
	}
	return path
}
