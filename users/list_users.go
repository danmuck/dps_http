package users

import (
	"context"

	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/bson"
)

func ListUsersT() ([]*User, error) {
	var users []*User
	cursor, err := MongoClient.Collection("users").Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	// temp debugging
	func(users []*User) {
		for _, user := range users {
			logs.Info("List of user: %+v", user)
		}
	}(users)

	return users, nil
}
