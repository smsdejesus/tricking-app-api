package services

import (
	"context"
	"errors"
	"fmt"

	"tricking-api/internal/models"
	"tricking-api/internal/repository"
)

// =============================================================================
// CUSTOM ERRORS FOR SERVICE LAYER
// =============================================================================

// ErrTrickNotFound indicates the requested trick doesn't exist
// We create service-layer errors separate from repository errors
// This allows us to change repository implementation without changing handlers
var ErrTrickNotFound = errors.New("trick not found")

// =============================================================================
// SERVICE INTERFACE
// =============================================================================

// TrickServiceInterface defines the contract for trick business operations
type TrickServiceInterface interface {
	GetSimpleTrickById(ctx context.Context, id string) (*models.TrickDetailResponse, error)
	GetFullDetailsTrickById(ctx context.Context, id string) (*models.TrickFullDetailsResponse, error)
	GetSimpleTricksList(ctx context.Context) ([]models.TrickSimpleResponse, error)
	GetLastModified(ctx context.Context) (int64, error)
	GetLastModifiedByID(ctx context.Context, id string) (int64, error)
}

// =============================================================================
// SERVICE IMPLEMENTATION
// =============================================================================

// TrickService implements TrickServiceInterface
type TrickService struct {
	// Services can depend on multiple repositories
	trickRepo repository.TrickRepositoryInterface
	videoRepo repository.VideoRepositoryInterface
}

// NewTrickService creates a new TrickService instance
// Accepts interfaces, not concrete types - this enables mocking for tests
func NewTrickService(trickRepo repository.TrickRepositoryInterface, videoRepo repository.VideoRepositoryInterface) *TrickService {
	return &TrickService{
		trickRepo: trickRepo,
		videoRepo: videoRepo,
	}
}

// GetSimpleTrickById retrieves basic trick details without videos
// "simple" endpoint
func (s *TrickService) GetSimpleTrickById(ctx context.Context, id string) (*models.TrickDetailResponse, error) {
	// Fetch trick from repository
	trick, err := s.trickRepo.GetByID(ctx, id)
	if err != nil {
		// Convert repository errors to service errors
		// This abstracts the data layer from the handler layer
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTrickNotFound
		}
		// Wrap unexpected errors with context
		return nil, fmt.Errorf("failed to get trick: %w", err)
	}

	// Convert model to response DTO
	// The handler doesn't need to know about this transformation
	response := trick.ToDetailResponse()
	return &response, nil
}

// GetFullDetailsTrickById retrieves full trick details WITH videos
func (s *TrickService) GetFullDetailsTrickById(ctx context.Context, id string) (*models.TrickFullDetailsResponse, error) {

	// Step 1: Get the trick
	trick, err := s.trickRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTrickNotFound
		}
		return nil, fmt.Errorf("failed to get trick: %w", err)
	}

	// Step 2: Get all videos for this trick
	videos, err := s.videoRepo.FindByTrickID(ctx, id)
	if err != nil {
		// We could decide to return the trick without videos on error
		// Business decision: should video fetch failure fail the whole request?
		// Here we choose to fail - adjust based on your requirements
		return nil, fmt.Errorf("failed to get videos for trick: %w", err)
	}

	// Step 3: Convert videos to response DTOs
	videoResponses := make([]models.VideoResponse, 0, len(videos))
	var featuredVideo *models.VideoResponse

	for _, video := range videos {
		vr := video.ToResponse()
		videoResponses = append(videoResponses, vr)

		// Track the featured video for convenience
		if video.IsFeatured {
			featuredVideo = &vr
			break
		}
	}

	// Step 4: Build the combined response
	response := &models.TrickFullDetailsResponse{
		TrickDetailResponse: trick.ToDetailResponse(),
		FeaturedVideo:       featuredVideo,
	}

	return response, nil
}

// GetSimpleTricksList retrieves a minimal list for dropdown menus
func (s *TrickService) GetSimpleTricksList(ctx context.Context) ([]models.TrickSimpleResponse, error) {
	// Call repository method
	tricks, err := s.trickRepo.FindSimpleList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tricks list: %w", err)
	}
	return tricks, nil
}

// GetLastModified returns the latest modification timestamp across all tricks
// Used for efficient ETag generation on list endpoints
func (s *TrickService) GetLastModified(ctx context.Context) (int64, error) {
	timestamp, err := s.trickRepo.GetLastModified(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get last modified timestamp: %w", err)
	}
	return timestamp, nil
}

// GetLastModifiedByID returns the modification timestamp for a specific trick
// Used for efficient ETag generation on individual trick endpoints
func (s *TrickService) GetLastModifiedByID(ctx context.Context, id string) (int64, error) {
	timestamp, err := s.trickRepo.GetLastModifiedByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrTrickNotFound
		}
		return 0, fmt.Errorf("failed to get last modified timestamp for trick: %w", err)
	}
	return timestamp, nil
}
