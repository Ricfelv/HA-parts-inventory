package parts

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"ha-parts-inventory/internal/database"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "parts.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewService(NewRepository(db))
}

func TestListSearchFilterAndSort(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	first, err := service.Create(ctx, Input{Type: "Discrete_Resistor", Value: "10kΩ", Package: "0603", MPN: "RC0603", Quantity: 4})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Create(ctx, Input{Type: "Discrete_Capacitor", Value: "10µF", Package: "0603", Quantity: 0})
	if err != nil {
		t.Fatal(err)
	}

	items, err := service.List(ctx, ListOptions{Query: "rc0603", InStockOnly: true, Sort: "value", Direction: "asc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != first {
		t.Fatalf("unexpected results: %#v", items)
	}
	items, err = service.List(ctx, ListOptions{InStockOnly: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d parts, want 2", len(items))
	}
}

func TestQuantityCannotBecomeNegative(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	id, err := service.Create(ctx, Input{Type: "Discrete_LED", Value: "Red", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AdjustQuantity(ctx, id, 3, false); err == nil {
		t.Fatal("expected a negative quantity error")
	}
	part, err := service.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if part.Quantity != 2 {
		t.Fatalf("quantity changed to %d", part.Quantity)
	}
	if err := service.AdjustQuantity(ctx, id, 2, false); err != nil {
		t.Fatal(err)
	}
	part, err = service.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if part.Quantity != 0 {
		t.Fatalf("quantity = %d, want 0", part.Quantity)
	}
}

func TestInputValidation(t *testing.T) {
	service := newTestService(t)
	_, err := service.Create(context.Background(), Input{Type: "Not a type", Value: "x", Quantity: 1})
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
