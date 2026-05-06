package trip

import "time"

type Trip struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	StartDate            *string    `json:"start_date"`
	EndDate              *string    `json:"end_date"`
	TotalDistanceKm      float64    `json:"total_distance_km"`
	EstimatedDurationSec int        `json:"estimated_duration_sec"`
	Status               string     `json:"status"`
	CoverPhotoURL        *string    `json:"cover_photo_url"`
	GpxURL               *string    `json:"gpx_url"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateRequest struct {
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	StartDate            *string `json:"start_date"`
	EndDate              *string `json:"end_date"`
	TotalDistanceKm      float64 `json:"total_distance_km"`
	EstimatedDurationSec int     `json:"estimated_duration_sec"`
}

type UpdateRequest struct {
	Title                *string  `json:"title"`
	Description          *string  `json:"description"`
	StartDate            *string  `json:"start_date"`
	EndDate              *string  `json:"end_date"`
	TotalDistanceKm      *float64 `json:"total_distance_km"`
	EstimatedDurationSec *int     `json:"estimated_duration_sec"`
	Status               *string  `json:"status"`
}

const FreeTierTripLimit = 3
