// =============================================================================
// FILE: internal/repository/combo_repository.go
// PURPOSE: Database operations for saved combos
// =============================================================================
//
// This handles user-saved combos. A combo is a sequence of tricks.
// The data model uses a junction table (combo_tricks) for the many-to-many
// relationship between combos and tricks.
//
// TABLE STRUCTURE (you'll need to create these):
//
// CREATE TABLE combos (
//     id BIGSERIAL PRIMARY KEY,
//     user_id UUID NOT NULL,
//     name TEXT NOT NULL,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
// );
//
// CREATE TABLE combo_tricks (
//     combo_id BIGINT REFERENCES combos(id) ON DELETE CASCADE,
//     trick_id INTEGER REFERENCES tricks(id),
//     position INTEGER NOT NULL,  -- Order in the combo
//     PRIMARY KEY (combo_id, trick_id, position)
// );
// =============================================================================

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tricking-api/internal/models"
)

// ComboRepositoryInterface defines the contract for combo data operations
type ComboRepositoryInterface interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Combo, error)
	GetTricksForCombo(ctx context.Context, comboID int64) ([]models.TrickSimpleResponse, error)
	Create(ctx context.Context, userID uuid.UUID, name string, trickIDs []int) (*models.Combo, error)
}

// ComboRepository implements ComboRepositoryInterface
type ComboRepository struct {
	pool *pgxpool.Pool
}

// NewComboRepository creates a new ComboRepository instance
func NewComboRepository(pool *pgxpool.Pool) *ComboRepository {
	return &ComboRepository{pool: pool}
}

// FindByUserID retrieves all combos for a specific user
func (r *ComboRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Combo, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM combos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query combos for user: %w", err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	combos, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Combo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect combo rows: %w", err)
	}

	return combos, nil
}

// Create saves a new combo with its tricks
// Uses a transaction to ensure atomic creation
func (r *ComboRepository) Create(ctx context.Context, userID uuid.UUID, name string, trickIDs []int) (*models.Combo, error) {
	// ==========================================================================
	// TRANSACTION EXAMPLE
	// ==========================================================================
	// A transaction ensures that either ALL operations succeed, or NONE do.
	// This prevents partial data (combo without tricks, or orphaned tricks).

	// Begin transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback - this is a no-op if we commit, but ensures cleanup on error
	defer tx.Rollback(ctx)

	// Insert the combo and get its ID
	// RETURNING id is a PostgreSQL feature that returns the generated ID
	var comboID int64
	var createdAt time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO combos (user_id, name) VALUES ($1, $2) RETURNING id, created_at`,
		userID, name,
	).Scan(&comboID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert combo: %w", err)
	}

	// Insert each trick in the combo
	for position, trickID := range trickIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO combo_tricks (combo_id, trick_id, position) VALUES ($1, $2, $3)`,
			comboID, trickID, position+1, // Position is 1-indexed
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert combo trick: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.Combo{
		ID:        comboID,
		UserID:    userID,
		Name:      name,
		CreatedAt: createdAt,
	}, nil
}
