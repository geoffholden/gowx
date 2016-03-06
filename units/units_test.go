// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package units

import (
	"math"
)

func floatEquals(a, b float64) bool {
	diff := math.Abs(a - b)
	a = math.Abs(a)
	b = math.Abs(b)
	m := math.Max(a, b)
	return diff <= m*1e-5
}
