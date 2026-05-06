package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("trip not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM trips WHERE user_id=$1 AND deleted_at IS NULL`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *Repository) ListByUser(ctx context.Context, userID string, page, limit int) ([]Trip, int, error) {
	offset := (page - 1) * limit

	var total int
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM trips WHERE user_id=$1 AND deleted_at IS NULL`,
		userID,
	).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, title, description,
		       start_date, end_date,
		       total_distance_km, estimated_duration_sec,
		       status, cover_photo_url, gpx_url,
		       created_at, updated_at
		FROM trips
		WHERE user_id=$1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list trips: %w", err)
	}
	defer rows.Close()

	var trips []Trip
	for rows.Next() {
		t, err := scanTrip(rows)
		if err != nil {
			return nil, 0, err
		}
		trips = append(trips, t)
	}
	return trips, total, nil
}

func (r *Repository) GetByID(ctx context.Context, id, userID string) (*Trip, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, title, description,
		       start_date, end_date,
		       total_distance_km, estimated_duration_sec,
		       status, cover_photo_url, gpx_url,
		       created_at, updated_at
		FROM trips
		WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL
	`, id, userID)

	t, err := scanTrip(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get trip: %w", err)
	}
	return &t, nil
}

func (r *Repository) Create(ctx context.Context, userID string, req CreateRequest) (*Trip, error) {
	var t Trip
	err := r.db.QueryRow(ctx, `
		INSERT INTO trips(user_id, title, description, start_date, end_date,
		                  total_distance_km, estimated_duration_sec)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, user_id, title, description,
		          start_date, end_date,
		          total_distance_km, estimated_duration_sec,
		          status, cover_photo_url, gpx_url,
		          created_at, updated_at
	`, userID, req.Title, req.Description, req.StartDate, req.EndDate,
		req.TotalDistanceKm, req.EstimatedDurationSec,
	).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description,
		&t.StartDate, &t.EndDate,
		&t.TotalDistanceKm, &t.EstimatedDurationSec,
		&t.Status, &t.CoverPhotoURL, &t.GpxURL,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create trip: %w", err)
	}
	return &t, nil
}

func (r *Repository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Trip, error) {
	existing, err := r.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Patch — sadece gönderilen alanları güncelle
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = req.EndDate
	}
	if req.TotalDistanceKm != nil {
		existing.TotalDistanceKm = *req.TotalDistanceKm
	}
	if req.EstimatedDurationSec != nil {
		existing.EstimatedDurationSec = *req.EstimatedDurationSec
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	var t Trip
	err = r.db.QueryRow(ctx, `
		UPDATE trips SET
			title=$1, description=$2, start_date=$3, end_date=$4,
			total_distance_km=$5, estimated_duration_sec=$6,
			status=$7, updated_at=$8
		WHERE id=$9 AND user_id=$10 AND deleted_at IS NULL
		RETURNING id, user_id, title, description,
		          start_date, end_date,
		          total_distance_km, estimated_duration_sec,
		          status, cover_photo_url, gpx_url,
		          created_at, updated_at
	`,
		existing.Title, existing.Description, existing.StartDate, existing.EndDate,
		existing.TotalDistanceKm, existing.EstimatedDurationSec,
		existing.Status, time.Now(),
		id, userID,
	).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description,
		&t.StartDate, &t.EndDate,
		&t.TotalDistanceKm, &t.EstimatedDurationSec,
		&t.Status, &t.CoverPhotoURL, &t.GpxURL,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update trip: %w", err)
	}
	return &t, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE trips SET deleted_at=NOW() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete trip: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// scanTrip pgx row → Trip
func scanTrip(row interface {
	Scan(dest ...any) error
}) (Trip, error) {
	var t Trip
	err := row.Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description,
		&t.StartDate, &t.EndDate,
		&t.TotalDistanceKm, &t.EstimatedDurationSec,
		&t.Status, &t.CoverPhotoURL, &t.GpxURL,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}
