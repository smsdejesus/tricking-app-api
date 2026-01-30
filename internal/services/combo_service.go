package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"tricking-api/internal/models"
	"tricking-api/internal/repository"
)

// CUSTOM ERRORS
var (
	ErrInsufficientTricks = errors.New("not enough tricks available for requested combo size")
	ErrInvalidComboSize   = errors.New("combo size must be at least 1")
)

type ComboServiceInterface interface {
	GenerateComboWithFilters(ctx context.Context, req models.ComboGenerateRequest) (*models.GeneratedComboResponse, error)
	GenerateSimpleCombo(ctx context.Context, size int) (*models.GeneratedComboResponse, error)
}

type ComboService struct {
	trickRepo repository.TrickRepositoryInterface
	rng       *rand.Rand // Random number generator for combo generation
}

// NewComboService creates a new ComboService instance
func NewComboService(trickRepo repository.TrickRepositoryInterface) *ComboService {
	return &ComboService{
		trickRepo: trickRepo,
		// Create a seeded random generator
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateComboWithFilters creates a new combo based on filters
// This is the "complicated" version with all filter options
func (s *ComboService) GenerateComboWithFilters(ctx context.Context, req models.ComboGenerateRequest) (*models.GeneratedComboResponse, error) {
	// ==========================================================================
	// VALIDATION
	// ==========================================================================
	if req.Size < 1 {
		return nil, ErrInvalidComboSize
	}

	// ==========================================================================
	// FETCH CANDIDATE TRICKS
	// ==========================================================================
	// First, get all tricks that match the filters
	filters := repository.TrickFilters{
		MaxDifficulty:   req.MaxDifficulty,
		CategoryIDs:     req.ExcludeCategoryIDs,
		ExcludeTrickIDs: req.ExcludeTrickIDs,
	}

	candidateTricks, err := s.trickRepo.FindByFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tricks for combo generation: %w", err)
	}

	// Check if we have enough tricks
	if len(candidateTricks) < req.Size {
		return nil, fmt.Errorf("%w: need %d tricks, only %d available",
			ErrInsufficientTricks, req.Size, len(candidateTricks))
	}

	// ==========================================================================
	// COMBO GENERATION ALGORITHM
	// ==========================================================================
	// This is where the business logic lives!
	//
	// Algorithm options you might implement:
	// 1. Random selection (simple)
	// 2. Weighted random (higher weight = more likely)
	// 3. Flow-based (consider landing_stance -> takeoff_stance compatibility)
	// 4. Difficulty progression (start easy, build up)
	// 5. Variety enforcement (no duplicate trick types in a row)

	selectedTricks := s.selectTricksWeighted(candidateTricks, req.Size)

	// ==========================================================================
	// BUILD RESPONSE
	// ==========================================================================
	return s.buildComboResponse(selectedTricks), nil
}

// GenerateSimpleCombo creates a combo based only on size (no filters)
// This is the "simple" version
func (s *ComboService) GenerateSimpleCombo(ctx context.Context, size int) (*models.GeneratedComboResponse, error) {
	if size < 1 {
		return nil, ErrInvalidComboSize
	}

	// Get all tricks (no filters)
	allTricks, err := s.trickRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tricks: %w", err)
	}

	if len(allTricks) < size {
		return nil, fmt.Errorf("%w: need %d tricks, only %d available",
			ErrInsufficientTricks, size, len(allTricks))
	}

	selectedTricks := s.selectTricksWeighted(allTricks, size)
	return s.buildComboResponse(selectedTricks), nil
}

// =============================================================================
// PRIVATE HELPER METHODS
// =============================================================================

// selectTricksWeighted selects n tricks using weighted random selection
// Tricks with higher weight are more likely to be selected
func (s *ComboService) selectTricksWeighted(candidates []models.Trick, count int) []models.Trick {

	// Make a copy to avoid modifying the original slice
	available := make([]models.Trick, len(candidates))
	copy(available, candidates)

	selected := make([]models.Trick, 0, count)

	for i := 0; i < count && len(available) > 0; i++ {
		// Calculate total weight
		totalWeight := int64(0)
		for _, trick := range available {
			// Ensure minimum weight of 1 to prevent tricks from being impossible to select
			weight := int64(trick.Weight)
			if weight < 1 {
				weight = 1
			}
			totalWeight += weight
		}

		// Pick random point in weight space
		target := s.rng.Int63n(totalWeight)

		// Find the trick at that point
		cumulative := int64(0)
		selectedIdx := 0
		for idx, trick := range available {
			weight := int64(trick.Weight)
			if weight < 1 {
				weight = 1
			}
			cumulative += weight
			if cumulative > target {
				selectedIdx = idx
				break
			}
		}

		// Add to selected and remove from available
		selected = append(selected, available[selectedIdx])
		// Remove by swapping with last element and shrinking slice
		available[selectedIdx] = available[len(available)-1]
		available = available[:len(available)-1]
	}

	return selected
}

// buildComboResponse creates the API response from selected tricks
func (s *ComboService) buildComboResponse(tricks []models.Trick) *models.GeneratedComboResponse {
	// Convert to simple responses
	trickResponses := make([]models.TrickSimpleResponse, 0, len(tricks))

	for _, trick := range tricks {
		trickResponses = append(trickResponses, trick.ToSimpleResponse())
	}

	return &models.GeneratedComboResponse{
		Tricks: trickResponses,
	}
}

// =============================================================================
// ALTERNATIVE SELECTION ALGORITHMS (for reference)
// =============================================================================

// selectTricksWithFlow considers stance compatibility for smoother combos
// This is more complex but creates more realistic combos
func (s *ComboService) selectTricksWithFlow(candidates []models.Trick, count int) []models.Trick {
	if len(candidates) == 0 || count == 0 {
		return []models.Trick{}
	}

	selected := make([]models.Trick, 0, count)
	available := make([]models.Trick, len(candidates))
	copy(available, candidates)

	// Pick first trick randomly (weighted)
	first := s.pickWeightedRandom(available)
	selected = append(selected, first)
	available = s.removeTrick(available, first.ID)

	// For subsequent tricks, prefer those where takeoff_stance matches previous landing_stance
	for i := 1; i < count && len(available) > 0; i++ {
		lastTrick := selected[i-1]

		// Find tricks with compatible stances
		compatible := s.filterCompatibleTricks(available, lastTrick.LandingStanceID)

		var nextTrick models.Trick
		if len(compatible) > 0 {
			// Pick from compatible tricks
			nextTrick = s.pickWeightedRandom(compatible)
		} else {
			// Fallback to any trick if no compatible ones
			nextTrick = s.pickWeightedRandom(available)
		}

		selected = append(selected, nextTrick)
		available = s.removeTrick(available, nextTrick.ID)
	}

	return selected
}

// pickWeightedRandom picks a single trick using weighted random selection
func (s *ComboService) pickWeightedRandom(tricks []models.Trick) models.Trick {
	if len(tricks) == 1 {
		return tricks[0]
	}

	totalWeight := int64(0)
	for _, t := range tricks {
		w := int64(t.Weight)
		if w < 1 {
			w = 1
		}
		totalWeight += w
	}

	target := s.rng.Int63n(totalWeight)
	cumulative := int64(0)

	for _, t := range tricks {
		w := int64(t.Weight)
		if w < 1 {
			w = 1
		}
		cumulative += w
		if cumulative > target {
			return t
		}
	}

	return tricks[len(tricks)-1] // Fallback
}

// filterCompatibleTricks returns tricks where takeoff matches the given landing stance
func (s *ComboService) filterCompatibleTricks(tricks []models.Trick, landingStanceID *int) []models.Trick {
	if landingStanceID == nil {
		return tricks // No landing stance = any trick works
	}

	compatible := make([]models.Trick, 0)
	for _, t := range tricks {
		// Trick is compatible if it has no takeoff requirement OR matches
		if t.TakeoffStanceID == nil || *t.TakeoffStanceID == *landingStanceID {
			compatible = append(compatible, t)
		}
	}
	return compatible
}

// removeTrick removes a trick from a slice by ID
func (s *ComboService) removeTrick(tricks []models.Trick, id int) []models.Trick {
	for i, t := range tricks {
		if t.ID == id {
			return append(tricks[:i], tricks[i+1:]...)
		}
	}
	return tricks
}
