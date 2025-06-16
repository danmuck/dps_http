package storage

import (
	"slices"

	"github.com/danmuck/dps_http/api/logs"
	"go.mongodb.org/mongo-driver/bson"
)

func CleanAndPrefix(filter any) bson.M {
	fm, ok := filter.(bson.M)
	if !ok {
		return bson.M{}
	}

	out := bson.M{}
	for key, val := range fm {
		if key == "key" {
			// special case for "key" to avoid prefixing
			logs.Log("[utils] allowing key %q without prefix", key)
			out[key] = val
			continue
		}
		if slices.Contains(allowed, key) {
			prefixed := "value." + key
			logs.Log("[utils] allowing %q → %q", key, prefixed)
			out[prefixed] = val
		}
	}
	return out
}
