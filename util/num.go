package util

import "math"

func Int32Min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func Distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}
