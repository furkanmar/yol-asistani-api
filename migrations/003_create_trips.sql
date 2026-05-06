CREATE TABLE IF NOT EXISTS trips (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                   TEXT NOT NULL,
    description             TEXT NOT NULL DEFAULT '',
    start_date              DATE,
    end_date                DATE,
    total_distance_km       FLOAT NOT NULL DEFAULT 0,
    estimated_duration_sec  INT NOT NULL DEFAULT 0,
    status                  TEXT NOT NULL DEFAULT 'draft',
    cover_photo_url         TEXT,
    gpx_url                 TEXT,
    deleted_at              TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trips_user_id ON trips(user_id);
