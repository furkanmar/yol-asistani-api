package route

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Service orchestrates OSRM routing with Redis cache.
type Service struct {
	osrm  *OSRMClient
	cache *RouteCache
}

func NewService(osrm *OSRMClient, cache *RouteCache) *Service {
	return &Service{osrm: osrm, cache: cache}
}

// ──────────────────────────────────────────────
// Alternatives — single origin→destination
// ──────────────────────────────────────────────

func (s *Service) Alternatives(ctx context.Context, req AlternativesRequest, tier string) (AlternativesResponse, error) {
	alts := MaxAlternatives(tier, req.Alternatives)

	key := alternativesKey(req.OriginLat, req.OriginLng, req.DestLat, req.DestLng, alts)

	if cached, ok := s.cache.GetRoutes(ctx, key); ok {
		return AlternativesResponse{Routes: cached}, nil
	}

	points := [][2]float64{
		{req.OriginLat, req.OriginLng},
		{req.DestLat, req.DestLng},
	}

	routes, err := s.osrm.RoutePoints(points, alts)
	if err != nil {
		return AlternativesResponse{}, fmt.Errorf("osrm: %w", err)
	}

	s.cache.SetRoutes(ctx, key, routes)
	return AlternativesResponse{Routes: routes}, nil
}

// ──────────────────────────────────────────────
// Segment — per-segment alternatives for N waypoints
// ──────────────────────────────────────────────

func (s *Service) Segment(ctx context.Context, req SegmentRequest, tier string) (SegmentResponse, error) {
	points, err := parseWaypoints(req.Waypoints)
	if err != nil {
		return SegmentResponse{}, fmt.Errorf("waypoints: %w", err)
	}
	if len(points) < 2 {
		return SegmentResponse{}, fmt.Errorf("need at least 2 waypoints")
	}

	// Pro+ only gets segment breakdown with alternatives
	alts := MaxAlternatives(tier, req.Alternatives)

	segments := make([]Segment, 0, len(points)-1)
	for i := 0; i < len(points)-1; i++ {
		from := points[i]
		to := points[i+1]

		key := segmentKey(from, to, alts)

		var routes []Route
		if cached, ok := s.cache.GetRoutes(ctx, key); ok {
			routes = cached
		} else {
			routes, err = s.osrm.RoutePoints([][2]float64{from, to}, alts)
			if err != nil {
				return SegmentResponse{}, fmt.Errorf("segment %d→%d: %w", i, i+1, err)
			}
			s.cache.SetRoutes(ctx, key, routes)
		}

		segments = append(segments, Segment{
			FromIndex: i,
			ToIndex:   i + 1,
			Routes:    routes,
		})
	}
	return SegmentResponse{Segments: segments}, nil
}

// ──────────────────────────────────────────────
// Parse helpers
// ──────────────────────────────────────────────

// parseWaypoints parses "lat1,lng1;lat2,lng2;..." into [][2]float64.
func parseWaypoints(raw string) ([][2]float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty waypoints")
	}
	parts := strings.Split(raw, ";")
	out := make([][2]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pair := strings.Split(p, ",")
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid pair %q", p)
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(pair[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("lat parse %q: %w", pair[0], err)
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(pair[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("lng parse %q: %w", pair[1], err)
		}
		out = append(out, [2]float64{lat, lng})
	}
	return out, nil
}
