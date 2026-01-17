// =============================================================================
// FILE: internal/services/category_service.go
// PURPOSE: Business logic for categories
// =============================================================================
//
// This is a simple service - categories don't have complex business logic.
// However, having a service layer provides:
// 1. Consistency with other entities
// 2. A place to add caching later
// 3. A place to add business logic if needed (e.g., sorting by popularity)
// =============================================================================

package services

import (
	"context"
	"fmt"

	"tricking-api/internal/models"
	"tricking-api/internal/repository"
)

// CategoryServiceInterface defines the contract for category operations
type CategoryServiceInterface interface {
	GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error)
}

// CategoryService implements CategoryServiceInterface
type CategoryService struct {
	categoryRepo repository.CategoryRepositoryInterface
}

// NewCategoryService creates a new CategoryService instance
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// GetAllCategories retrieves all categories for the UI dropdown
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Convert to response DTOs
	responses := make([]models.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		responses = append(responses, cat.ToResponse())
	}

	return responses, nil
}
