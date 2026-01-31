package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tricking-api/internal/models"
)

// CategoryRepositoryInterface defines the contract for category data operations
type CategoryRepositoryInterface interface {
	FindAll(ctx context.Context) ([]models.Category, error)
}

// CategoryRepository implements CategoryRepositoryInterface
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new CategoryRepository instance
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// FindAll retrieves all categories
// This is used to populate dropdown menus in the UI
func (r *CategoryRepository) FindAll(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, parent_id
		FROM trick_data.categories
		ORDER BY parent_id DESC, name ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect category rows: %w", err)
	}

	return categories, nil
}
