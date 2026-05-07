package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OSRMClient calls the OSRM route API.
type OSRMClient struct {
	baseURL string
	http    *http.Client
}

func NewOSRMClient(baseURL string) *OSRMClient {
	return &OSRMClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ──────────────────────────────────────────────
// OSRM wire types
// ──────────────────────────────────────────────

type osrmResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64        `json:"distance"` // metres
	Duration float64        `json:"duration"` // seconds
	Geometry osrmGeometry   `json:"geometry"`
	Legs     []osrmLeg      `json:"legs"`
}

type osrmGeometry struct {
	Type        string       `json:"type"`
	Coordinates [][2]float64 `json:"coordinates"` // [lng, lat]
}

type osrmLeg struct {
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
}

// ──────────────────────────────────────────────
// Route — single origin→destination
// ──────────────────────────────────────────────

// RoutePoints calls OSRM with up to `alts` alternatives.
// Points are [lat, lng]; OSRM expects lng,lat in URL.
func (c *OSRMClient) RoutePoints(points [][2]float64, alts int) ([]Route, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("need at least 2 points")
	}

	coords := make([]string, len(points))
	for i, p := range points {
		coords[i] = fmt.Sprintf("%.6f,%.6f", p[1], p[0]) // lng,lat
	}
	coordStr := strings.Join(coords, ";")

	url := fmt.Sprintf(
		"%s/route/v1/driving/%s?alternatives=%d&geometries=geojson&overview=full&steps=false",
		c.baseURL, coordStr, alts,
	)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("osrm request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("osrm read body: %w", err)
	}

	var osrmResp osrmResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, fmt.Errorf("osrm parse: %w", err)
	}
	if osrmResp.Code != "Ok" {
		return nil, fmt.Errorf("osrm error code: %s", osrmResp.Code)
	}

	routes := make([]Route, 0, len(osrmResp.Routes))
	for i, r := range osrmResp.Routes {
		// Convert [lng,lat] → [lat,lng] for curviness + ViaPoints
		coords := lngLatToLatLng(r.Geometry.Coordinates)

		// Rebuild geometry with lat,lng for consistency
		geoOut := map[string]any{
			"type":        r.Geometry.Type,
			"coordinates": r.Geometry.Coordinates, // keep original GeoJSON [lng,lat]
		}

		routes = append(routes, Route{
			Index:          i,
			DistanceKm:     r.Distance / 1000.0,
			DurationSec:    r.Duration,
			CurvinessScore: CalculateCurviness(coords),
			Geometry:       geoOut,
			ViaPoints:      sampleViaPoints(coords, 10),
		})
	}
	return routes, nil
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

// lngLatToLatLng swaps coordinate pairs [lng,lat] → [lat,lng].
func lngLatToLatLng(coords [][2]float64) [][2]float64 {
	out := make([][2]float64, len(coords))
	for i, c := range coords {
		out[i] = [2]float64{c[1], c[0]}
	}
	return out
}

// sampleViaPoints returns up to n evenly-spaced points from coords.
func sampleViaPoints(coords [][2]float64, n int) [][2]float64 {
	if len(coords) == 0 {
		return nil
	}
	if len(coords) <= n {
		return coords
	}
	out := make([][2]float64, n)
	step := float64(len(coords)-1) / float64(n-1)
	for i := range n {
		idx := int(float64(i) * step)
		out[i] = coords[idx]
	}
	return out
}
