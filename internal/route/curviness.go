package route

import "math"

// CalculateCurviness computes a 0–1 curviness score from a coordinate sequence.
// Algorithm: sum absolute bearing changes between consecutive segments,
// normalise by (totalBearingChange / distKm). 100 deg/km → score 1.0.
func CalculateCurviness(coords [][2]float64) float64 {
	if len(coords) < 3 {
		return 0
	}

	totalBearingChange := 0.0
	for i := 1; i < len(coords)-1; i++ {
		b1 := bearing(coords[i-1], coords[i])
		b2 := bearing(coords[i], coords[i+1])
		diff := math.Abs(b2 - b1)
		if diff > 180 {
			diff = 360 - diff
		}
		totalBearingChange += diff
	}

	distKm := totalDistanceKm(coords)
	if distKm == 0 {
		return 0
	}
	degPerKm := totalBearingChange / distKm
	return math.Min(degPerKm/100.0, 1.0)
}

// bearing returns the initial bearing (degrees, 0–360) from a to b.
func bearing(a, b [2]float64) float64 {
	lat1 := toRad(a[0])
	lat2 := toRad(b[0])
	dLng := toRad(b[1] - a[1])

	y := math.Sin(dLng) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLng)
	deg := toDeg(math.Atan2(y, x))
	return math.Mod(deg+360, 360)
}

// haversineKm returns great-circle distance in km between two [lat,lng] points.
func haversineKm(a, b [2]float64) float64 {
	const R = 6371.0
	dLat := toRad(b[0] - a[0])
	dLng := toRad(b[1] - a[1])
	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	h := sinDLat*sinDLat + math.Cos(toRad(a[0]))*math.Cos(toRad(b[0]))*sinDLng*sinDLng
	return 2 * R * math.Asin(math.Sqrt(h))
}

func totalDistanceKm(coords [][2]float64) float64 {
	total := 0.0
	for i := 1; i < len(coords); i++ {
		total += haversineKm(coords[i-1], coords[i])
	}
	return total
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }
func toDeg(rad float64) float64 { return rad * 180 / math.Pi }
