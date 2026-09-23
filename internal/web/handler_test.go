package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"ha-parts-inventory/internal/database"
	"ha-parts-inventory/internal/parts"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "parts.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	handler, err := NewHandler(parts.NewService(parts.NewRepository(db)))
	if err != nil {
		t.Fatal(err)
	}
	return handler.Routes()
}

func TestCreateAndInventory(t *testing.T) {
	handler := testHandler(t)
	form := "type=Discrete_Resistor&value=10k&package=0603&description=1%25&mpn=RC0603&quantity=7"
	request := httptest.NewRequest(http.MethodPost, "/parts", strings.NewReader(form))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d", response.Code)
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "10k") {
		t.Fatal("inventory did not contain created part")
	}
}

func TestIngressPathsAreUsed(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Ingress-Path", "/api/hassio_ingress/example")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, `href="/api/hassio_ingress/example/static/style.css"`) {
		t.Fatal("stylesheet did not use ingress path")
	}
	if !strings.Contains(body, `action="/api/hassio_ingress/example/"`) {
		t.Fatal("filter form did not use ingress path")
	}
}

func TestInventoryDefaultsToInStockOnly(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "parts.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := parts.NewService(parts.NewRepository(db))
	if _, err := service.Create(t.Context(), parts.Input{Type: "Discrete_Resistor", Value: "10k", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(t.Context(), parts.Input{Type: "Discrete_Capacitor", Value: "10uF", Quantity: 0}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if strings.Contains(response.Body.String(), "10uF") {
		t.Fatal("out-of-stock part appeared in the default inventory view")
	}
}
