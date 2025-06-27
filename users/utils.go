package users

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func VerifyUsername(username string) bool {
	err := service.storage.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(nil)
	return err != mongo.ErrNoDocuments
}

func VerifyEmail(email string) bool {
	err := service.storage.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(nil)
	return err != mongo.ErrNoDocuments
}

func VerifyNew(username, email string) bool {
	filter := bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
	}

	var existingUser User
	err := service.storage.Collection("users").FindOne(context.Background(), filter).Decode(&existingUser)
	return err == mongo.ErrNoDocuments
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashed, password string) bool {
	check := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return check == nil
}
