package parts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("part not found")

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, options ListOptions) ([]Part, error) {
	return s.repository.List(ctx, options)
}
func (s *Service) Get(ctx context.Context, id int64) (Part, error) {
	part, err := s.repository.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Part{}, ErrNotFound
	}
	return part, err
}
func (s *Service) Create(ctx context.Context, input Input) (int64, error) {
	if err := validate(input); err != nil {
		return 0, err
	}
	return s.repository.Create(ctx, normalize(input))
}
func (s *Service) Update(ctx context.Context, id int64, input Input) error {
	if err := validate(input); err != nil {
		return err
	}
	err := s.repository.Update(ctx, id, normalize(input))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func (s *Service) AdjustQuantity(ctx context.Context, id int64, amount int, add bool) error {
	if amount < 0 {
		return ValidationError{"Adjustment must be zero or greater."}
	}
	err := s.repository.AdjustQuantity(ctx, id, amount, add)
	if errors.Is(err, sql.ErrNoRows) {
		if _, getErr := s.Get(ctx, id); errors.Is(getErr, ErrNotFound) {
			return ErrNotFound
		}
		return ValidationError{"Removing that amount would make quantity negative."}
	}
	return err
}
func validate(input Input) error {
	if !validType(input.Type) {
		return ValidationError{"Please select a valid part type."}
	}
	if strings.TrimSpace(input.Value) == "" {
		return ValidationError{"Value is required."}
	}
	if input.Quantity < 0 {
		return ValidationError{"Quantity must be zero or greater."}
	}
	return nil
}
func validType(value string) bool {
	for _, kind := range Types {
		if value == kind {
			return true
		}
	}
	return false
}
func normalize(input Input) Input {
	input.Type = strings.TrimSpace(input.Type)
	input.Value = strings.TrimSpace(input.Value)
	input.Package = strings.TrimSpace(input.Package)
	input.Description = strings.TrimSpace(input.Description)
	input.MPN = strings.TrimSpace(input.MPN)
	return input
}
