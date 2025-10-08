package utils

import (
	"math"
)

const (
	EarthRadiusKm = 6371.0
	SpeedKmH      = 40.0
)

func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = EarthRadiusKm * 1000
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func EstimateTimeMinutes(totalDistanceMeters float64) int {
	km := totalDistanceMeters / 1000.0
	hours := km / SpeedKmH
	minutes := hours * 60
	return int(math.Round(minutes))
}
