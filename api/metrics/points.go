package metrics

import "sort"

// Point represents a single data point in the time series for user metrics.
// it matches the format of the users_over_time map and the typescript type
type Point struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// MapTimestampToInt64Points sorts a map of timestamps to counts into a slice of Points.
func MapTimestampToInt64Points(m map[string]int64) []Point {
	points := make([]Point, 0, len(m))
	for k, v := range m {
		points = append(points, Point{
			Timestamp: k,
			Count:     v,
		})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp < points[j].Timestamp
	})

	return points
}
