package util

import "math"

func Int32Min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func Distance(x1, y1, x2, y2 float32) float32 {
	return float32(math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2)))
}
