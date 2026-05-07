package route

// ──────────────────────────────────────────────
// Request types
// ──────────────────────────────────────────────

type AlternativesRequest struct {
	OriginLat    float64 `query:"origin_lat"`
	OriginLng    float64 `query:"origin_lng"`
	DestLat      float64 `query:"dest_lat"`
	DestLng      float64 `query:"dest_lng"`
	Alternatives int     `query:"alternatives"`
}

type SegmentRequest struct {
	// "lat1,lng1;lat2,lng2;lat3,lng3"
	Waypoints    string `query:"waypoints"`
	Alternatives int    `query:"alternatives"`
}

// ──────────────────────────────────────────────
// Response types
// ──────────────────────────────────────────────

type Route struct {
	Index          int          `json:"index"`
	DistanceKm     float64      `json:"distance_km"`
	DurationSec    float64      `json:"duration_sec"`
	CurvinessScore float64      `json:"curviness_score"`
	Geometry       any          `json:"geometry"` // GeoJSON LineString
	ViaPoints      [][2]float64 `json:"via_points"`
}

type AlternativesResponse struct {
	Routes []Route `json:"routes"`
}

type Segment struct {
	FromIndex int     `json:"from_index"`
	ToIndex   int     `json:"to_index"`
	Routes    []Route `json:"routes"`
}

type SegmentResponse struct {
	Segments []Segment `json:"segments"`
}

// ──────────────────────────────────────────────
// Tier limits
// ──────────────────────────────────────────────

var tierAlternativesLimit = map[string]int{
	"free":     1,
	"pro":      3,
	"pro_plus": 5,
}

// MaxAlternatives returns max allowed alternatives count for tier.
// Clamps requested value to tier ceiling.
func MaxAlternatives(tier string, requested int) int {
	limit, ok := tierAlternativesLimit[tier]
	if !ok {
		limit = 1
	}
	if requested <= 0 || requested > limit {
		return limit
	}
	return requested
}
