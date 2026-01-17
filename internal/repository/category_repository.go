// =============================================================================
// FILE: internal/repository/category_repository.go
// PURPOSE: Database operations for trick categories
// =============================================================================
//
// Categories help users filter tricks. In your current schema, it looks like
// the `flip_id` column in tricks might reference a categories/flips table.
//
// You may need to create a categories table:
//
// CREATE TABLE categories (
//     id SERIAL PRIMARY KEY,
//     name TEXT NOT NULL,
//     type TEXT  -- e.g., 'flip', 'kick', 'twist', 'transition'
// );
// =============================================================================

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
	GetByID(ctx context.Context, id int) (*models.Category, error)
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
		SELECT id, name, COALESCE(type, '') as type
		FROM categories
		ORDER BY name ASC
	`
	// COALESCE handles NULL values - if type is NULL, use empty string
	// This prevents NULL scan issues

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

// GetByID retrieves a single category by its ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, name, COALESCE(type, '') as type
		FROM categories
		WHERE id = $1
	`

	var category models.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Type,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get category by ID %d: %w", id, err)
	}

	return &category, nil
}
