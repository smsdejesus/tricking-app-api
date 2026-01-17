// =============================================================================
// FILE: internal/services/user_service.go
// PURPOSE: Business logic for user-related operations
// =============================================================================

package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"tricking-api/internal/models"
	"tricking-api/internal/repository"
)

// UserServiceInterface defines the contract for user operations
type UserServiceInterface interface {
	GetUserCombos(ctx context.Context, userID uuid.UUID) ([]models.ComboResponse, error)
	// Add more user-related methods as needed:
	// GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)
	// UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs models.UserPreferences) error
}

// UserService implements UserServiceInterface
type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserCombos retrieves all saved combos for a user with their tricks
func (s *UserService) GetUserCombos(ctx context.Context, userID uuid.UUID) ([]models.ComboResponse, error) {
	// Get the user's combos
	combos, err := s.userRepo.GetCombosByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user combos: %w", err)
	}

	// Build response with tricks for each combo
	responses := make([]models.ComboResponse, 0, len(combos))

	for _, combo := range combos {
		// Get tricks for this combo
		tricks, err := s.userRepo.GetComboTricks(ctx, combo.ID)
		if err != nil {
			// Log error but continue - don't fail the whole request for one bad combo
			// In production, use a proper logger
			fmt.Printf("Warning: failed to get tricks for combo %d: %v\n", combo.ID, err)
			tricks = []models.TrickSimpleResponse{} // Empty slice instead of nil
		}

		responses = append(responses, models.ComboResponse{
			ID:        combo.ID,
			Name:      combo.Name,
			Tricks:    tricks,
			CreatedAt: combo.CreatedAt,
		})
	}

	return responses, nil
}
