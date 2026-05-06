CREATE TABLE IF NOT EXISTS waypoints (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id          UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    position         INT NOT NULL DEFAULT 0,
    label            TEXT NOT NULL DEFAULT '',
    lat              FLOAT NOT NULL,
    lng              FLOAT NOT NULL,
    location         GEOMETRY(Point, 4326),
    notes            TEXT NOT NULL DEFAULT '',
    stay_duration_min INT NOT NULL DEFAULT 0,
    poi_id           TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_waypoints_trip_id  ON waypoints(trip_id);
CREATE INDEX IF NOT EXISTS idx_waypoints_location ON waypoints USING GIST(location);
