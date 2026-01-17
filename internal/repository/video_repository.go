// =============================================================================
// FILE: internal/repository/video_repository.go
// PURPOSE: Database operations for trick videos
// =============================================================================
//
// This repository handles all video-related database operations.
// It's kept separate from TrickRepository to follow Single Responsibility Principle.
// =============================================================================

package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tricking-api/internal/models"
)

// VideoRepositoryInterface defines the contract for video data operations
type VideoRepositoryInterface interface {
	FindByTrickID(ctx context.Context, trickID int) ([]models.TrickVideo, error)
	GetFeaturedByTrickID(ctx context.Context, trickID int) (*models.TrickVideo, error)
}

// VideoRepository implements VideoRepositoryInterface
type VideoRepository struct {
	pool *pgxpool.Pool
}

// NewVideoRepository creates a new VideoRepository instance
func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

// FindByTrickID retrieves all videos for a specific trick
func (r *VideoRepository) FindByTrickID(ctx context.Context, trickID int) ([]models.TrickVideo, error) {
	query := `
		SELECT 
			id, trick_id, video_url, thumbnail_url,
			uploaded_by, performer_user_id, performer_name,
			is_featured, created_at
		FROM trick_videos
		WHERE trick_id = $1
		ORDER BY is_featured DESC, created_at DESC
	`
	// ORDER BY is_featured DESC puts featured videos first
	// Then by created_at DESC to show newest videos first

	rows, err := r.pool.Query(ctx, query, trickID)
	if err != nil {
		return nil, fmt.Errorf("failed to query videos for trick %d: %w", trickID, err)
	}

	// pgx.CollectRows handles iteration, scanning, and closing rows automatically
	videos, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.TrickVideo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect video rows: %w", err)
	}

	return videos, nil
}

// GetFeaturedByTrickID retrieves the featured video for a trick
// Returns nil (not error) if no featured video exists
func (r *VideoRepository) GetFeaturedByTrickID(ctx context.Context, trickID int) (*models.TrickVideo, error) {
	query := `
		SELECT 
			id, trick_id, video_url, thumbnail_url,
			uploaded_by, performer_user_id, performer_name,
			is_featured, created_at
		FROM trick_videos
		WHERE trick_id = $1 AND is_featured = true
		LIMIT 1
	`

	var video models.TrickVideo
	err := r.pool.QueryRow(ctx, query, trickID).Scan(
		&video.ID,
		&video.TrickID,
		&video.VideoURL,
		&video.ThumbnailURL,
		&video.UploadedBy,
		&video.PerformerUserID,
		&video.PerformerName,
		&video.IsFeatured,
		&video.CreatedAt,
	)

	if err != nil {
		// No featured video is not an error - just return nil
		// This is different from TrickRepository where not finding a trick IS an error
		// Design decision: missing featured video is expected, missing trick is not
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get featured video for trick %d: %w", trickID, err)
	}

	return &video, nil
}
