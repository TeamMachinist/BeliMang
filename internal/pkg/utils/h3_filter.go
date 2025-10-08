package utils

import (
	"github.com/uber/h3-go/v4"
)

const (
	H3_RES_FOR_FILTER  = 10     // resolusi H3
	EDGE_LENGTH_METERS = 104.8  // panjang sisi heksagon di res 10 (meter)
	SAFE_ACCEPT_M      = 2500.0 // batas bawah: pasti ≤3000m
	SAFE_REJECT_M      = 3500.0 // batas atas: pasti >3000m
)

// konvert lat long ke h3 cell di resolusi H3_RES_FOR_FILTER
func LatLonToH3(lat, lng float64) (h3.Cell, error) {
	return h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, H3_RES_FOR_FILTER)
}

// H3GridDistanceMeters returns estimasi jarak dalam meter menggunakan H3 grid distance.
// Returns -1 if distance cannot be computed.
func H3GridDistanceMeters(a, b h3.Cell) float64 {
	if a == b {
		return 0
	}
	gridDist, err := h3.GridDistance(a, b)
	if err != nil {
		return -1
	}
	return float64(gridDist) * EDGE_LENGTH_METERS
}
