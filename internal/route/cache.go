package route

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheTTL = 24 * time.Hour

// RouteCache wraps Redis for route result caching.
type RouteCache struct {
	rdb *redis.Client
}

func NewRouteCache(redisURL string) (*RouteCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis parse url: %w", err)
	}
	rdb := redis.NewClient(opts)

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RouteCache{rdb: rdb}, nil
}

// ──────────────────────────────────────────────
// Cache key helpers
// ──────────────────────────────────────────────

// alternativesKey builds a cache key from 5-decimal coords.
// 5 decimal places ≈ 1 m precision — good enough to avoid spurious misses.
func alternativesKey(originLat, originLng, destLat, destLng float64, alts int) string {
	return fmt.Sprintf("route:%.5f:%.5f:%.5f:%.5f:%d",
		originLat, originLng, destLat, destLng, alts)
}

func segmentKey(from, to [2]float64, alts int) string {
	return fmt.Sprintf("seg:%.5f:%.5f:%.5f:%.5f:%d",
		from[0], from[1], to[0], to[1], alts)
}

// ──────────────────────────────────────────────
// Get / Set
// ──────────────────────────────────────────────

func (rc *RouteCache) GetRoutes(ctx context.Context, key string) ([]Route, bool) {
	val, err := rc.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var routes []Route
	if err := json.Unmarshal(val, &routes); err != nil {
		return nil, false
	}
	return routes, true
}

func (rc *RouteCache) SetRoutes(ctx context.Context, key string, routes []Route) {
	b, err := json.Marshal(routes)
	if err != nil {
		return
	}
	// Fire-and-forget — cache failure is not critical
	_ = rc.rdb.SetEx(ctx, key, b, cacheTTL).Err()
}
