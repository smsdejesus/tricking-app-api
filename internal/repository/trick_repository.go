// =============================================================================
// FILE: internal/repository/trick_repository.go
// PURPOSE: Database operations for tricks - the "data access layer"
// =============================================================================
//
// REPOSITORY PATTERN:
// The repository pattern abstracts database operations behind an interface.
// This means:
// 1. Your service layer doesn't know about SQL - it just calls repository methods
// 2. You can swap PostgreSQL for MySQL by creating a new repository implementation
// 3. You can easily mock the repository for unit testing
//
// NAMING CONVENTIONS:
// - Repository suffix: TrickRepository, VideoRepository, etc.
// - Method names describe the data operation: GetByID, FindAll, Create, Update, Delete
// - "Get" typically returns one item or error
// - "Find" or "List" typically returns multiple items
// - "Create" inserts new data
// - "Update" modifies existing data
// - "Delete" removes data
// =============================================================================

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tricking-api/internal/models"
)

// =============================================================================
// CUSTOM ERRORS
// =============================================================================
// Define custom errors that can be checked by the service layer
// This allows services to handle "not found" differently from database errors

// ErrNotFound indicates the requested resource doesn't exist
var ErrNotFound = errors.New("resource not found")

// =============================================================================
// INTERFACE DEFINITION
// =============================================================================
// Defining an interface allows for easy mocking in tests and swapping implementations

// TrickRepositoryInterface defines the contract for trick data operations
// NAMING: Interfaces in Go often end with "er" (Reader, Writer) or describe capability
// For repositories, "Interface" suffix is common for clarity
type TrickRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*models.Trick, error)
	FindAll(ctx context.Context) ([]models.Trick, error)
	FindSimpleList(ctx context.Context) ([]models.TrickSimpleResponse, error)
	FindByFilters(ctx context.Context, filters TrickFilters) ([]models.Trick, error)
}

// TrickFilters holds optional filters for querying tricks
type TrickFilters struct {
	MinDifficulty   *int64
	MaxDifficulty   *int64
	CategoryIDs     []int
	ExcludeTrickIDs []int
	Limit           *int
}

// =============================================================================
// REPOSITORY IMPLEMENTATION
// =============================================================================

// TrickRepository implements TrickRepositoryInterface using PostgreSQL
type TrickRepository struct {
	// pool is the database connection pool
	// Using lowercase (unexported) because external packages shouldn't access it directly
	pool *pgxpool.Pool
}

// NewTrickRepository creates a new TrickRepository instance
// NAMING: "New" + StructName is the Go convention for constructors
func NewTrickRepository(pool *pgxpool.Pool) *TrickRepository {
	return &TrickRepository{pool: pool}
}

// GetByID retrieves a single trick by its ID
// Returns ErrNotFound if the trick doesn't exist
func (r *TrickRepository) GetByID(ctx context.Context, id int) (*models.Trick, error) {
	// SQL query to fetch a single trick
	// $1 is a placeholder for the first parameter (prevents SQL injection)
	// NEVER use fmt.Sprintf to build queries with user input!
	query := `
		SELECT 
			id, name, description, difficulty, execution_notes,
			created_by, creator_name, created_at, updated_at,
			takeoff_stance_id, landing_stance_id, flip_id, rotation, weight
		FROM trick_data.tricks
		WHERE id = $1
	`

	// Create an empty Trick to scan results into
	var trick models.Trick

	// QueryRow is used when expecting exactly one row
	// Scan maps columns to struct fields in ORDER - must match SELECT order!
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&trick.ID,
		&trick.Name,
		&trick.Description,
		&trick.Difficulty,
		&trick.ExecutionNotes,
		&trick.CreatedBy, // Can be NULL, so we use *uuid.UUID
		&trick.CreatorName,
		&trick.CreatedAt,
		&trick.UpdatedAt,
		&trick.TakeoffStanceID, // Can be NULL, so we use *int
		&trick.LandingStanceID,
		&trick.FlipID,
		&trick.Rotation,
		&trick.Weight,
	)
	fmt.Println("Retrieved trick:", trick)
	fmt.Println("Retrieved ERROR:", err)
	if err != nil {
		// Check if it's a "no rows" error
		if errors.Is(err, pgx.ErrNoRows) {
			// Return our custom error so the service layer knows it's "not found"
			return nil, ErrNotFound
		}
		// Wrap other errors with context
		return nil, fmt.Errorf("failed to get trick by ID %d: %w", id, err)
	}

	return &trick, nil
}

// FindAll retrieves all tricks from the database
func (r *TrickRepository) FindAll(ctx context.Context) ([]models.Trick, error) {
	query := `
		SELECT 
			id, name, description, difficulty, execution_notes,
			created_by, creator_name, created_at, updated_at,
			takeoff_stance_id, landing_stance_id, flip_id, rotation, weight
		FROM trick_data.tricks
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tricks: %w", err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	// RowToStructByName maps columns to struct fields using db tags
	tricks, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Trick])
	if err != nil {
		return nil, fmt.Errorf("failed to collect trick rows: %w", err)
	}

	return tricks, nil
}

// FindSimpleList retrieves a minimal list of tricks for dropdown menus
// This is more efficient than FindAll when you only need ID and name
func (r *TrickRepository) FindSimpleList(ctx context.Context) ([]models.TrickSimpleResponse, error) {
	// Only select the columns we need - more efficient!
	query := `
		SELECT id, name
		FROM trick_data.tricks
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tricks simple list: %w", err)
	}

	// pgx.CollectRows with RowToStructByPos for simple DTOs without db tags
	tricks, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.TrickSimpleResponse])
	if err != nil {
		return nil, fmt.Errorf("failed to collect trick simple rows: %w", err)
	}

	return tricks, nil
}

// FindByFilters retrieves tricks matching the given filters
// This is used by the combo generation algorithm
func (r *TrickRepository) FindByFilters(ctx context.Context, filters TrickFilters) ([]models.Trick, error) {
	// ==========================================================================
	// DYNAMIC QUERY BUILDING
	// ==========================================================================
	// We build the query dynamically based on which filters are provided.
	// This is a common pattern for search/filter functionality.
	//
	// IMPORTANT: We use parameterized queries ($1, $2, etc.) to prevent SQL injection.
	// Never concatenate user input directly into SQL strings!

	// Base query
	query := `
		SELECT 
			id, name, description, difficulty, execution_notes,
			created_by, creator_name, created_at, updated_at,
			takeoff_stance_id, landing_stance_id, flip_id, rotation, weight
		FROM trick_data.tricks
		WHERE 1=1
	`
	// "WHERE 1=1" is a trick that makes it easier to append AND clauses
	// because every condition can start with "AND"

	// args holds the parameter values in order ($1, $2, etc.)
	args := make([]interface{}, 0)
	argPosition := 1 // Tracks which $N we're on

	// Add difficulty filters if provided
	if filters.MinDifficulty != nil {
		query += fmt.Sprintf(" AND difficulty >= $%d", argPosition)
		args = append(args, *filters.MinDifficulty)
		argPosition++
	}

	if filters.MaxDifficulty != nil {
		query += fmt.Sprintf(" AND difficulty <= $%d", argPosition)
		args = append(args, *filters.MaxDifficulty)
		argPosition++
	}

	// Add category filter if provided
	// This assumes you have a category_id column or a junction table
	// Adjust based on your actual schema
	if len(filters.CategoryIDs) > 0 {
		query += fmt.Sprintf(" AND flip_id = ANY($%d)", argPosition)
		args = append(args, filters.CategoryIDs)
		argPosition++
	}

	// Exclude specific tricks
	if len(filters.ExcludeTrickIDs) > 0 {
		query += fmt.Sprintf(" AND id != ALL($%d)", argPosition)
		args = append(args, filters.ExcludeTrickIDs)
		argPosition++
	}

	// Add ordering - we order by weight for combo generation
	// Higher weight = more likely to be selected
	query += " ORDER BY weight DESC, RANDOM()"

	// Add limit if specified
	if filters.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argPosition)
		args = append(args, *filters.Limit)
	}

	// Execute the query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tricks with filters: %w", err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	tricks, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Trick])
	if err != nil {
		return nil, fmt.Errorf("failed to collect filtered trick rows: %w", err)
	}

	return tricks, nil
}

// =============================================================================
// ALTERNATIVE: Using pgx.CollectRows for cleaner code (pgx v5)
// =============================================================================
// pgx v5 provides helper functions that reduce boilerplate:
//
// func (r *TrickRepository) FindAllClean(ctx context.Context) ([]models.Trick, error) {
//     query := `SELECT id, name, ... FROM tricks`
//
//     rows, err := r.pool.Query(ctx, query)
//     if err != nil {
//         return nil, err
//     }
//
//     // pgx.CollectRows handles the iteration and scanning
//     return pgx.CollectRows(rows, pgx.RowToStructByName[models.Trick])
// }
//
// Note: RowToStructByName requires your struct fields to have `db:"column_name"` tags
// =============================================================================
