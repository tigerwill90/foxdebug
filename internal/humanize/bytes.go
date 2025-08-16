// The code in this package is derivative of https://github.com/dustin/go-humanize (all credit to Dustin Sallings).
// Mount of this source code is governed by a MIT License that can be found
// at https://github.com/dustin/go-humanize/blob/master/LICENSE.

package humanize

import (
	"fmt"
	"math"
)

// Bytes produces a human-readable representation of an SI size.
func Bytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanizeBytes(s, 1000, sizes)
}

func humanizeBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
