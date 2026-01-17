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
	GetTrickSimple(ctx context.Context, id int) (*models.TrickDetailResponse, error)
	GetTrickDictionary(ctx context.Context, id int) (*models.TrickDictionaryResponse, error)
	GetTricksList(ctx context.Context) ([]models.TrickSimpleResponse, error)
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
// Notice we accept interfaces, not concrete types - this enables mocking for tests
func NewTrickService(trickRepo *repository.TrickRepository, videoRepo *repository.VideoRepository) *TrickService {
	return &TrickService{
		trickRepo: trickRepo,
		videoRepo: videoRepo,
	}
}

// GetTrickSimple retrieves basic trick details without videos
// "simple" endpoint
func (s *TrickService) GetTrickSimple(ctx context.Context, id int) (*models.TrickDetailResponse, error) {
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

// GetTrickDictionary retrieves full trick details WITH videos
// "complicated/dictionary" endpoint
func (s *TrickService) GetTrickDictionary(ctx context.Context, id int) (*models.TrickDictionaryResponse, error) {
	// ==========================================================================
	// ORCHESTRATION EXAMPLE
	// ==========================================================================
	// This method combines data from TWO repositories (tricks + videos)
	// The handler doesn't need to know these are separate database queries

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
		if video.IsFeatured && featuredVideo == nil {
			featuredVideo = &vr
		}
	}

	// Step 4: Build the combined response
	response := &models.TrickDictionaryResponse{
		TrickDetailResponse: trick.ToDetailResponse(),
		Videos:              videoResponses,
		FeaturedVideo:       featuredVideo,
	}

	return response, nil
}

// GetTricksList retrieves a minimal list for dropdown menus
func (s *TrickService) GetTricksList(ctx context.Context) ([]models.TrickSimpleResponse, error) {
	// Direct pass-through - no business logic needed
	// But having this in a service still provides:
	// 1. A consistent interface for handlers
	// 2. A place to add caching later
	// 3. A place to add filtering/sorting logic later
	tricks, err := s.trickRepo.FindSimpleList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tricks list: %w", err)
	}
	return tricks, nil
}

// =============================================================================
// OPTIONAL: Caching example
// =============================================================================
//
//
// type TrickService struct {
//     trickRepo    repository.TrickRepositoryInterface
//     videoRepo    repository.VideoRepositoryInterface
//     cache        *cache.Cache  // Some caching library
//     cacheTTL     time.Duration
// }
//
// func (s *TrickService) GetTricksListCached(ctx context.Context) ([]models.TrickSimpleResponse, error) {
//     // Try cache first
//     if cached, ok := s.cache.Get("tricks_list"); ok {
//         return cached.([]models.TrickSimpleResponse), nil
//     }
//
//     // Cache miss - fetch from database
//     tricks, err := s.trickRepo.FindSimpleList(ctx)
//     if err != nil {
//         return nil, err
//     }
//
//     // Store in cache
//     s.cache.Set("tricks_list", tricks, s.cacheTTL)
//
//     return tricks, nil
// }
// =============================================================================
