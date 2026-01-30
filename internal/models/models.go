package models

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// DATABASE MODELS - These match PostgreSQL table structures
// =============================================================================

// Trick represents a row in the "tricks" table
// STRUCT TAGS: The `db:"column_name"` tags tell pgx which column to map to which field
// The `json:"field_name"` tags control JSON serialization for API responses
type Trick struct {
	// ID is the primary key
	ID int `db:"id" json:"id"`

	// URL-friendly unique identifier for the trick
	Slug string `db:"slug" json:"slug"`

	// Name is the trick name (e.g., "Backflip", "540 Kick")
	Name string `db:"name" json:"name"`

	// Description explains what the trick is (nullable)
	Description *string `db:"description" json:"description,omitempty"`

	// Difficulty is a numeric rating (e.g., 1-10)
	// Using pointer (*int64) to allow NULL values from database
	Difficulty *int64 `db:"difficulty" json:"difficulty,omitempty"`

	// ExecutionNotes provides tips on how to perform the trick (nullable)
	ExecutionNotes *string `db:"execution_notes" json:"execution_notes,omitempty"`

	// CreatedBy is the UUID of the user who created this trick entry
	// Using pointer (*uuid.UUID) allows this to be null in the database
	CreatedBy *uuid.UUID `db:"created_by" json:"-"` // json:"-" hides this from API responses

	// CreatorName is denormalized for display without joins (nullable)
	CreatorName *string `db:"creator_name" json:"creator_name,omitempty"`

	// CreatedAt is when this record was created (has default but nullable)
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`

	// UpdatedAt is when this record was last modified (has default but nullable)
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`

	// TakeoffStanceID links to the stance table (foreign key)
	// Pointer allows null values
	TakeoffStanceID *int `db:"takeoff_stance_id" json:"takeoff_stance_id,omitempty"`

	// LandingStanceID links to the stance table (foreign key)
	LandingStanceID *int `db:"landing_stance_id" json:"landing_stance_id,omitempty"`

	// FlipID categorizes the type of flip (foreign key to flips/categories table)
	FlipID *int `db:"flip_id" json:"flip_id,omitempty"`

	// Rotation is the degrees of rotation (e.g., 180, 360, 540) - nullable
	Rotation *int `db:"rotation" json:"rotation,omitempty"`

	// Weight is used for combo generation algorithm (affects selection probability)
	Weight int16 `db:"weight" json:"weight"`
}

// TrickVideo represents a row in the "trick_videos" table
type TrickVideo struct {
	// ID is the primary key (bigint in PostgreSQL = int64 in Go)
	ID int64 `db:"id" json:"id"`

	// TrickID links this video to a trick (foreign key)
	TrickID int `db:"trick_id" json:"trick_id"`

	// VideoURL is the URL to the video file
	VideoURL string `db:"video_url" json:"video_url"`

	// ThumbnailURL is the URL to the video thumbnail image
	ThumbnailURL string `db:"thumbnail_url" json:"thumbnail_url"`

	// UploadedBy is the UUID of the user who uploaded this video
	UploadedBy uuid.UUID `db:"uploaded_by" json:"-"`

	// PerformerUserID is the UUID of the user performing in the video (if registered)
	// Pointer allows null (performer might not have an account)
	PerformerUserID *uuid.UUID `db:"performer_user_id" json:"-"`

	// PerformerName is the name of the person performing the trick
	PerformerName string `db:"performer_name" json:"performer_name"`

	// IsFeatured indicates if this is the primary/featured video for the trick
	IsFeatured bool `db:"is_featured" json:"is_featured"`

	// CreatedAt is when this video was uploaded
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Category represents a trick category (for filtering)
// NEED to create this table if it doesn't exist
type Category struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	ParentID *int   `db:"parent_id" json:"parent_id,omitempty"`
}

// Combo represents a saved combo by a user
// NEED to create this table if it doesn't exist
type Combo struct {
	ID        int64     `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"-"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ComboTrick represents the many-to-many relationship between combos and tricks
// This is a junction/join table
type ComboTrick struct {
	ComboID  int64 `db:"combo_id" json:"combo_id"`
	TrickID  int   `db:"trick_id" json:"trick_id"`
	Position int   `db:"position" json:"position"` // Order in the combo (1st, 2nd, 3rd trick)
}

// =============================================================================
// API RESPONSE DTOs - These are what we send back to clients
// =============================================================================

// TrickSimpleResponse is a minimal trick representation for dropdowns/lists
type TrickSimpleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TrickDetailResponse is the full trick data without videos
// Used for the "simple" version of the trick detail endpoint
type TrickDetailResponse struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	Difficulty      *int64     `json:"difficulty,omitempty"`
	ExecutionNotes  *string    `json:"execution_notes,omitempty"`
	CreatorName     *string    `json:"creator_name,omitempty"`
	TakeoffStanceID *int       `json:"takeoff_stance_id,omitempty"`
	LandingStanceID *int       `json:"landing_stance_id,omitempty"`
	Rotation        *int       `json:"rotation,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

// VideoResponse is the video data for API responses
type VideoResponse struct {
	ID            int64     `json:"id"`
	VideoURL      string    `json:"video_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	PerformerName string    `json:"performer_name"`
	IsFeatured    bool      `json:"is_featured"`
	CreatedAt     time.Time `json:"created_at"`
}

// TrickFullDetailsResponse is the "complicated" version with video
// This is like a dictionary page for the trick with all available information
type TrickFullDetailsResponse struct {
	// Embed TrickDetailResponse to include all its fields
	// This is Go's composition pattern - avoids repeating fields
	TrickDetailResponse

	// FeaturedVideo is the primary video (convenience field)
	// Pointer allows null if no featured video exists
	FeaturedVideo *VideoResponse `json:"featured_video,omitempty"`
}

// ComboResponse represents a saved combo with its tricks
type ComboResponse struct {
	ID        int64                 `json:"id"`
	Name      string                `json:"name"`
	Tricks    []TrickSimpleResponse `json:"tricks"` // Ordered list of tricks
	CreatedAt time.Time             `json:"created_at"`
}

// GeneratedComboResponse represents a newly generated combo
type GeneratedComboResponse struct {
	Tricks []TrickSimpleResponse `json:"tricks"`
}

// CategoryResponse is for the categories list endpoint
type CategoryResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id,omitempty"`
}

// =============================================================================
// API REQUEST DTOs - These are what clients send to us
// =============================================================================

// ComboGenerateRequest contains filters for combo generation
// STRUCT TAGS:
// - `json:"field"` for JSON parsing
// - `form:"field"` for query parameters (GET requests)
// - `binding:"required"` makes the field mandatory (Gin validation)
type ComboGenerateRequest struct {
	// Size is the number of tricks in the combo (REQUIRED)
	Size int `json:"size" form:"size" binding:"required,min=1,max=10"`

	// The following filters are OPTIONAL (no binding:"required")

	// MaxDifficulty limits individual trick difficulty
	MaxDifficulty *int64 `json:"max_difficulty" form:"max_difficulty" binding:"omitempty,min=1"`

	// CategoryIDs filters tricks to specific categories
	// In query string: ?category_ids=1&category_ids=2&category_ids=3
	ExcludeCategoryIDs []int `json:"category_ids" form:"category_ids"`

	// TrickIDs specifies exact tricks to include (for partial customization)
	TrickIDs []int `json:"trick_ids" form:"trick_ids"`

	// ExcludeTrickIDs specifies tricks to never include
	ExcludeTrickIDs []int `json:"exclude_trick_ids" form:"exclude_trick_ids"`
}

// ComboGenerateSimpleRequest only requires size (no filters)
type ComboGenerateSimpleRequest struct {
	Size int `json:"size" form:"size" binding:"required,min=1,max=10"`
}

// =============================================================================
// HELPER METHODS - Convert between models and DTOs
// =============================================================================

// ToSimpleResponse converts a Trick model to TrickSimpleResponse DTO
// This is a method on Trick (receiver is t *Trick)
func (t *Trick) ToSimpleResponse() TrickSimpleResponse {
	return TrickSimpleResponse{
		ID:   t.ID,
		Name: t.Name,
	}
}

// ToDetailResponse converts a Trick model to TrickDetailResponse DTO
func (t *Trick) ToDetailResponse() TrickDetailResponse {
	return TrickDetailResponse{
		ID:              t.ID,
		Name:            t.Name,
		Description:     t.Description,
		Difficulty:      t.Difficulty,
		ExecutionNotes:  t.ExecutionNotes,
		CreatorName:     t.CreatorName,
		TakeoffStanceID: t.TakeoffStanceID,
		LandingStanceID: t.LandingStanceID,
		Rotation:        t.Rotation,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

// ToResponse converts a TrickVideo model to VideoResponse DTO
func (v *TrickVideo) ToResponse() VideoResponse {
	return VideoResponse{
		ID:            v.ID,
		VideoURL:      v.VideoURL,
		ThumbnailURL:  v.ThumbnailURL,
		PerformerName: v.PerformerName,
		IsFeatured:    v.IsFeatured,
		CreatedAt:     v.CreatedAt,
	}
}

// ToResponse converts a Category model to CategoryResponse DTO
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:       c.ID,
		Name:     c.Name,
		ParentID: c.ParentID,
	}
}
