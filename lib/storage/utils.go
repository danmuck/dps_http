package storage

import (
	"slices"

	"github.com/danmuck/dps_http/lib/logs"
	"go.mongodb.org/mongo-driver/bson"
)

// @NOTE this is kinda janky way to do this
// i may rehtink this approach
// //

// forbidden access to user data
// * not used for now
// var forbidden []string = []string{
// 	"password_hash", "token", "created_at", "updated_at",
// 	"bio", "avatar_url", "email",
// }

// allowed access to user data
// checked explicitly against this slice
var allowed []string = []string{
	"username", "roles",
}

// helper function to prefix a key with "value."
func Prefix(key string) string {
	return "value." + key
}

// helper function to check if a key is allowed
// and prefix them so they conform to MongoDB schema
func CleanAndPrefix(filter any) bson.M {
	fm, ok := filter.(bson.M)
	if !ok {
		return nil
	}

	out := bson.M{}
	for key, val := range fm {
		if key == "key" {
			// special case for "key" to avoid prefixing
			logs.Debug("allowing key %q without prefix", key)
			out[key] = val
			continue
		}
		if slices.Contains(allowed, key) {
			logs.Debug("allowing %q → %q", key, Prefix(key))
			out[Prefix(key)] = val
		}
	}
	return out
}
