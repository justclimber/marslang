package interpereter

import (
	"math"
)

func distance(x1, x2, y1, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	ds := (dx * dx) + (dy * dy)

	return math.Sqrt(ds)
}

func angle(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1

	return angleOrigin(dx, dy)
}

func angleOrigin(x, y float64) float64 {
	atan := math.Atan(y / x)
	if math.IsNaN(atan) {
		return 0.
	}
	switch {
	case x < 0 && y >= 0:
		return math.Pi + atan
	case x < 0 && y < 0:
		return math.Pi + atan
	case x > 0 && y < 0:
		return math.Pi*2 + atan
	default:
		return atan
	}
}
