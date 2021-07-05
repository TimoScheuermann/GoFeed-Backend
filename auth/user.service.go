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

/**
 * Returns the collection for users
 */
func collection() *mongo.Collection {
	return database.Database.Collection("user")
}

/**
 * This function takes an id and returns additional information
 * of an user
 */
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// read parameters defined in the controller
	params := mux.Vars(r)
	// get user by id, by extracting id parameter from params
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

/**
 * This functions takes an Id and returns the corresponding user
 * nil, if user with given id doesnt exist
 */
func GetUserById(userId string) *User {

	// convert id to objectId
	oid, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		fmt.Printf("Invalid ObjectID %v\n", userId)
		return nil
	}

	// Find user by id
	var user User
	err = collection().FindOne(ctx.Background(), bson.M{"_id": oid}).Decode(&user)

	if err != nil {
		fmt.Printf("User not found... (%v) %v \n", userId, err)
		return nil
	}

	return &user
}

/**
 * This function gets called by handleOAuthCallback() in auth.service.go
 */
func UserSignedIn(gothUser goth.User) *User {
	// Get user or nil, by given providerId and Provider
	user := getUserByProvider(gothUser.UserID, gothUser.Provider)

	if user == nil {
		// user doesnt exist => register
		return registerUser(gothUser)
	} else {
		// user does exist = update (e.g Name, Picture)
		return updateUser(user, gothUser)
	}
}

/**
 * This functions takes the providerId & provider and returns the corresponding user
 * nil, if user with given providerId & provider doesnt exist
 */
func getUserByProvider(providerId string, provider string) *User {

	// Find user and parse to var user
	var user User
	err := collection().FindOne(ctx.Background(), bson.M{"providerId": providerId, "provider": provider}).Decode(&user)

	if err != nil {
		return nil
	}

	return &user
}

/**
 * This function gets called by UserSignedIn in user.service.go.
 * It creates a new database entry for the given user
 */
func registerUser(gothUser goth.User) *User {
	// get current timestamp in milliseconds
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	// create user object, stored in database
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

	// store object in database
	result, _ := collection().InsertOne(ctx.Background(), user)

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		// return created user object
		return GetUserById(oid.Hex())
	}

	fmt.Println("Returning nil...")
	return nil
}

/**
 * This function gets called by UserSignedIn in user.service.go.
 * It updates an existing database entry for the given user
 */
func updateUser(user *User, gothUser goth.User) *User {
	// get current timestamp in milliseconds
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	fmt.Println("Update User")

	// update user's avatar, name and latest login timestamp
	collection().UpdateByID(ctx.Background(), user.UserID, bson.M{"$set": bson.M{
		"avatar":     gothUser.AvatarURL,
		"last_login": tUnixMilli,
		"name":       gothUser.Name,
	}})

	return GetUserById(user.UserID.Hex())
}
