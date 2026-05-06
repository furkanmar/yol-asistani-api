package waypoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("waypoint not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByTrip(ctx context.Context, tripID string) ([]Waypoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, trip_id, position, label, lat, lng,
		       notes, stay_duration_min, poi_id, created_at
		FROM waypoints
		WHERE trip_id=$1
		ORDER BY position ASC
	`, tripID)
	if err != nil {
		return nil, fmt.Errorf("list waypoints: %w", err)
	}
	defer rows.Close()

	var wps []Waypoint
	for rows.Next() {
		var w Waypoint
		if err := rows.Scan(
			&w.ID, &w.TripID, &w.Position, &w.Label, &w.Lat, &w.Lng,
			&w.Notes, &w.StayDurationMin, &w.PoiID, &w.CreatedAt,
		); err != nil {
			return nil, err
		}
		wps = append(wps, w)
	}
	return wps, nil
}

func (r *Repository) GetByID(ctx context.Context, id, tripID string) (*Waypoint, error) {
	var w Waypoint
	err := r.db.QueryRow(ctx, `
		SELECT id, trip_id, position, label, lat, lng,
		       notes, stay_duration_min, poi_id, created_at
		FROM waypoints
		WHERE id=$1 AND trip_id=$2
	`, id, tripID).Scan(
		&w.ID, &w.TripID, &w.Position, &w.Label, &w.Lat, &w.Lng,
		&w.Notes, &w.StayDurationMin, &w.PoiID, &w.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get waypoint: %w", err)
	}
	return &w, nil
}

func (r *Repository) Create(ctx context.Context, tripID string, req CreateRequest) (*Waypoint, error) {
	var w Waypoint
	err := r.db.QueryRow(ctx, `
		INSERT INTO waypoints(trip_id, position, label, lat, lng,
		                      notes, stay_duration_min, poi_id,
		                      location)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,
		       ST_SetSRID(ST_MakePoint($5,$4),4326))
		RETURNING id, trip_id, position, label, lat, lng,
		          notes, stay_duration_min, poi_id, created_at
	`, tripID, req.Position, req.Label, req.Lat, req.Lng,
		req.Notes, req.StayDurationMin, req.PoiID,
	).Scan(
		&w.ID, &w.TripID, &w.Position, &w.Label, &w.Lat, &w.Lng,
		&w.Notes, &w.StayDurationMin, &w.PoiID, &w.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create waypoint: %w", err)
	}
	return &w, nil
}

func (r *Repository) Update(ctx context.Context, id, tripID string, req UpdateRequest) (*Waypoint, error) {
	existing, err := r.GetByID(ctx, id, tripID)
	if err != nil {
		return nil, err
	}

	if req.Label != nil {
		existing.Label = *req.Label
	}
	if req.Lat != nil {
		existing.Lat = *req.Lat
	}
	if req.Lng != nil {
		existing.Lng = *req.Lng
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.StayDurationMin != nil {
		existing.StayDurationMin = *req.StayDurationMin
	}
	if req.Position != nil {
		existing.Position = *req.Position
	}

	var w Waypoint
	err = r.db.QueryRow(ctx, `
		UPDATE waypoints SET
			position=$1, label=$2, lat=$3, lng=$4,
			notes=$5, stay_duration_min=$6,
			location=ST_SetSRID(ST_MakePoint($4,$3),4326)
		WHERE id=$7 AND trip_id=$8
		RETURNING id, trip_id, position, label, lat, lng,
		          notes, stay_duration_min, poi_id, created_at
	`,
		existing.Position, existing.Label, existing.Lat, existing.Lng,
		existing.Notes, existing.StayDurationMin,
		id, tripID,
	).Scan(
		&w.ID, &w.TripID, &w.Position, &w.Label, &w.Lat, &w.Lng,
		&w.Notes, &w.StayDurationMin, &w.PoiID, &w.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update waypoint: %w", err)
	}
	return &w, nil
}

func (r *Repository) Delete(ctx context.Context, id, tripID string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM waypoints WHERE id=$1 AND trip_id=$2`,
		id, tripID,
	)
	if err != nil {
		return fmt.Errorf("delete waypoint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Reorder(ctx context.Context, tripID string, items []ReorderItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`UPDATE waypoints SET position=$1 WHERE id=$2 AND trip_id=$3`,
			item.Position, item.ID, tripID,
		)
		if err != nil {
			return fmt.Errorf("reorder waypoint %s: %w", item.ID, err)
		}
	}

	return tx.Commit(ctx)
}
