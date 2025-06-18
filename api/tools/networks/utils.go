package networks

import (
	"math"
	"strconv"
	"strings"
)

// default formatting constants helper
func FormatB(value float64) string {
	return FormatBits(value, 3, 2)
}

// FormatSize converts “value” into the largest human-readable unit so that
// 1 ≤ integer-part ≤ maxIntDigits, rounds to decDigits, trims trailing zeros,
// and appends the unit suffix.
func FormatBits(value float64, maxIntDigits, decDigits int) string {
	type unit struct {
		name string
		size float64
	}
	units := []unit{
		{"PB", PB}, {"Tb", Tb}, {"GB", GB}, {"Mb", Mb},
		{"MB", MB}, {"KB", KB}, {"Kb", Kb}, {"B", Byte}, {"b", Bit},
	}

	// pick the first unit where v = value/size ≥ 1 and integer-digits(v) ≤ maxIntDigits
	for _, u := range units {
		if value >= u.size {
			v := value / u.size
			if integerDigits(v) <= maxIntDigits {
				return formatFloat(v, decDigits) + " " + u.name
			}
		}
	}

	// fallback: just bits
	v := value / Bit
	return formatFloat(v, decDigits) + "b"
}

// integerDigits returns the count of digits left of the decimal in |v|.
// (e.g. v=0.5→1, v=12.3→2, v=1234→4)
func integerDigits(v float64) int {
	v = math.Abs(v)
	if v < 1 {
		return 1
	}
	return int(math.Floor(math.Log10(v))) + 1
}

// formatFloat produces a string with exactly decDigits places, then
// trims any trailing “0”s and a trailing “.” if present.
func formatFloat(v float64, decDigits int) string {
	s := strconv.FormatFloat(v, 'f', decDigits, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
