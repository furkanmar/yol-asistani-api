package waypoint

import "time"

type Waypoint struct {
	ID              string    `json:"id"`
	TripID          string    `json:"trip_id"`
	Position        int       `json:"position"`
	Label           string    `json:"label"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	Notes           string    `json:"notes"`
	StayDurationMin int       `json:"stay_duration_min"`
	PoiID           *string   `json:"poi_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateRequest struct {
	Position        int     `json:"position"`
	Label           string  `json:"label"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	Notes           string  `json:"notes"`
	StayDurationMin int     `json:"stay_duration_min"`
	PoiID           *string `json:"poi_id"`
}

type UpdateRequest struct {
	Label           *string  `json:"label"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
	Notes           *string  `json:"notes"`
	StayDurationMin *int     `json:"stay_duration_min"`
	Position        *int     `json:"position"`
}

type ReorderItem struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}
