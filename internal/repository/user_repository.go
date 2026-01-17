package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tricking-api/internal/models"
)

// UserRepositoryInterface defines the contract for user data operations
type UserRepositoryInterface interface {
	GetCombosByUserID(ctx context.Context, userID uuid.UUID) ([]models.Combo, error)
	GetComboTricks(ctx context.Context, comboID int64) ([]models.TrickSimpleResponse, error)
	// GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	// GetPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreferences, error)
}

// UserRepository implements UserRepositoryInterface
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetCombosByUserID retrieves all combos for a specific user
func (r *UserRepository) GetCombosByUserID(ctx context.Context, userID uuid.UUID) ([]models.Combo, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM combos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user combos: %w", err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	combos, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Combo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect combo rows: %w", err)
	}

	return combos, nil
}

// GetComboTricks retrieves all tricks for a specific combo, ordered by position
func (r *UserRepository) GetComboTricks(ctx context.Context, comboID int64) ([]models.TrickSimpleResponse, error) {
	query := `
		SELECT t.id, t.name
		FROM combo_tricks ct
		JOIN tricks t ON ct.trick_id = t.id
		WHERE ct.combo_id = $1
		ORDER BY ct.position ASC
	`

	rows, err := r.pool.Query(ctx, query, comboID)
	if err != nil {
		return nil, fmt.Errorf("failed to query combo tricks: %w", err)
	}

	// pgx.CollectRows with RowTo for simple structs without db tags
	tricks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[models.TrickSimpleResponse])
	if err != nil {
		return nil, fmt.Errorf("failed to collect trick rows: %w", err)
	}

	// Convert from []*TrickSimpleResponse to []TrickSimpleResponse
	result := make([]models.TrickSimpleResponse, len(tricks))
	for i, t := range tricks {
		result[i] = *t
	}

	return result, nil
}
