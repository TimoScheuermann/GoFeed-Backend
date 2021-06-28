package auth

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"gofeed-go/database"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func collection() *mongo.Collection {
	return database.Database.Collection("user")
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user := GetUserById(params["id"])

	if user == nil {
		json.NewEncoder(w).Encode(Exception{Message: "User not found"})
		return
	}

	type UserInfo struct {
		UserID primitive.ObjectID `json:"id"`
		Name   string             `json:"name"`
		Avatar string             `json:"avatar"`
	}

	json.NewEncoder(w).Encode(UserInfo{user.UserID, user.Name, user.Avatar})
}

func GetUserById(userId string) *User {
	var user User
	oid, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		fmt.Printf("Invalid ObjectID %v\n", userId)
		return nil
	}

	err = collection().FindOne(ctx.Background(), bson.M{"_id": oid}).Decode(&user)

	if err != nil {
		fmt.Printf("User not found... (%v) %v \n", userId, err)
		return nil
	}

	return &user
}

func UserSignedIn(gothUser goth.User) *User {
	user := getUserByProvider(gothUser.UserID, gothUser.Provider)

	if user == nil {
		return registerUser(gothUser)
	} else {
		return updateUser(user, gothUser)
	}
}

func getUserByProvider(providerId string, provider string) *User {
	var user User
	err := collection().FindOne(ctx.Background(), bson.M{"providerId": providerId, "provider": provider}).Decode(&user)

	if err != nil {
		return nil
	}

	return &user
}

func registerUser(gothUser goth.User) *User {
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	var user = User{
		ProviderID:  gothUser.UserID,
		Provider:    gothUser.Provider,
		Name:        gothUser.Name,
		Avatar:      gothUser.AvatarURL,
		Group:       "user",
		MemberSince: tUnixMilli,
		LastLogin:   tUnixMilli,
	}

	fmt.Println("Register User")

	result, _ := collection().InsertOne(ctx.Background(), user)

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return GetUserById(oid.Hex())
	}

	fmt.Println("Returning nil...")
	return nil
}

func updateUser(user *User, gothUser goth.User) *User {
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	fmt.Println("Update User")

	collection().UpdateByID(ctx.Background(), user.UserID, bson.M{"$set": bson.M{
		"avatar":     gothUser.AvatarURL,
		"last_login": tUnixMilli,
		"name":       gothUser.Name,
	}})

	return GetUserById(user.UserID.Hex())
}
