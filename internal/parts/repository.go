package parts

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, options ListOptions) ([]Part, error) {
	query := `SELECT id, type, value, package, description, mpn, quantity, updated_at FROM parts`
	var conditions []string
	var args []any
	if options.InStockOnly {
		conditions = append(conditions, "quantity > 0")
	}
	if options.Query != "" {
		conditions = append(conditions, `(LOWER(type) LIKE ? OR LOWER(value) LIKE ? OR LOWER(package) LIKE ? OR LOWER(description) LIKE ? OR LOWER(mpn) LIKE ?)`)
		needle := "%" + strings.ToLower(options.Query) + "%"
		for range 5 {
			args = append(args, needle)
		}
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY " + sortColumn(options.Sort) + " " + sortDirection(options.Direction) + ", id ASC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list parts: %w", err)
	}
	defer rows.Close()
	var result []Part
	for rows.Next() {
		part, err := scanPart(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, part)
	}
	return result, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id int64) (Part, error) {
	part, err := scanPart(r.db.QueryRowContext(ctx, `SELECT id, type, value, package, description, mpn, quantity, updated_at FROM parts WHERE id = ?`, id))
	if err != nil {
		return Part{}, err
	}
	return part, nil
}

func (r *Repository) Create(ctx context.Context, input Input) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO parts(type, value, package, description, mpn, quantity, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, input.Type, input.Value, input.Package, input.Description, input.MPN, input.Quantity, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("create part: %w", err)
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, id int64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE parts SET type = ?, value = ?, package = ?, description = ?, mpn = ?, quantity = ?, updated_at = ? WHERE id = ?`, input.Type, input.Value, input.Package, input.Description, input.MPN, input.Quantity, time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("update part: %w", err)
	}
	return requireRow(result)
}

func (r *Repository) AdjustQuantity(ctx context.Context, id int64, amount int, add bool) error {
	if add {
		result, err := r.db.ExecContext(ctx, `UPDATE parts SET quantity = quantity + ?, updated_at = ? WHERE id = ?`, amount, time.Now().UTC().Format(time.RFC3339), id)
		if err != nil {
			return fmt.Errorf("add quantity: %w", err)
		}
		return requireRow(result)
	}
	result, err := r.db.ExecContext(ctx, `UPDATE parts SET quantity = quantity - ?, updated_at = ? WHERE id = ? AND quantity >= ?`, amount, time.Now().UTC().Format(time.RFC3339), id, amount)
	if err != nil {
		return fmt.Errorf("remove quantity: %w", err)
	}
	return requireRow(result)
}

type scanner interface{ Scan(...any) error }

func scanPart(row scanner) (Part, error) {
	var part Part
	var updated string
	if err := row.Scan(&part.ID, &part.Type, &part.Value, &part.Package, &part.Description, &part.MPN, &part.Quantity, &updated); err != nil {
		return Part{}, err
	}
	parsed, err := time.Parse(time.RFC3339, updated)
	if err == nil {
		part.UpdatedAt = parsed
	}
	return part, nil
}

func requireRow(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func sortColumn(sort string) string {
	columns := map[string]string{"type": "type", "value": "value", "package": "package", "description": "description", "mpn": "mpn", "quantity": "quantity"}
	if column, ok := columns[sort]; ok {
		return column
	}
	return "id"
}

func sortDirection(direction string) string {
	if strings.EqualFold(direction, "desc") {
		return "DESC"
	}
	return "ASC"
}
