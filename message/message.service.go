package message

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"gofeed-go/auth"
	"gofeed-go/database"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
 * Returns the collection for messages
 */
func collection() *mongo.Collection {
	return database.Database.Collection("message")
}

/**
 * This function sends an unprocessable entity exception to the requester
 */
func printError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(Exception{message})
}

/**
 * This functions takes an Id and returns the corresponding message
 * nil, if message with given id doesnt exist
 */
func getMessageById(messageId string) *Message {
	// convert id to objectId
	oid, err := primitive.ObjectIDFromHex(messageId)

	if err != nil {
		fmt.Printf("Invalid ObjectID %v\n", messageId)
		return nil
	}

	// Find message by id
	var message Message
	err = collection().FindOne(ctx.Background(), bson.M{"_id": oid}).Decode(&message)

	if err != nil {
		fmt.Printf("Message not found... (%v) %v \n", messageId, err)
		return nil
	}

	return &message
}

/**
 * This function returns every message stored in the database
 * the requester can define additional queries (limit, skip) to
 * optimise the desired iutput
 */
func getMessages(w http.ResponseWriter, req *http.Request) {
	// Options for MongoDB Query
	options := options.Find()

	// Queries defined in the request
	query := req.URL.Query()
	qLimit, qSkip := query.Get("limit"), query.Get("skip")

	// Check if Limit Query exists
	if len(qLimit) > 0 {
		// Convert limitQuery to int64 and set options if no error occures
		if limit, err := strconv.ParseInt(qLimit, 10, 64); err == nil {
			options.SetLimit(limit)
		}
	}

	// Check if Skip Query exists
	if len(qSkip) > 0 {
		// Convert skipQuery to int64 and set options if no error occures
		if skip, err := strconv.ParseInt(qSkip, 10, 64); err == nil {
			options.SetSkip(skip)
		}
	}

	// find every dataset in the database
	cursor, err := collection().Find(ctx.Background(), bson.M{}, options)

	if err != nil {
		printError(w, err.Error())
		return
	}

	// close cursor after code has been fully executed
	defer cursor.Close(ctx.Background())

	// define array to store database results
	messages := []Message{}

	// iterate through results
	for cursor.Next(ctx.Background()) {

		// decode every single result
		var message Message
		cursor.Decode(&message)

		// append decoded result to array
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		printError(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(messages)
}

/**
 * This function takes an id and returns additional information
 * of a message
 */
func getMessage(w http.ResponseWriter, req *http.Request) {
	// read parameters defined in the controller
	params := mux.Vars(req)
	// get message by id, by extracting id parameter from params
	message := getMessageById(params["id"])

	if message == nil {
		printError(w, "Message not found")
		return
	}

	json.NewEncoder(w).Encode(message)
}

/**
 * This function reads the request's body and stores it in the database
 */
func postMessage(w http.ResponseWriter, req *http.Request) {
	// decode body into new message object
	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	// create a new validator
	validate := validator.New()
	// validate struct (validation defined in message/types.go (validate:"..."))
	// for additional information, check out: https://github.com/go-playground/validator
	err := validate.Struct(message)

	if err != nil {
		printError(w, "Bitte gib eine Nachricht ein")
		return
	}

	// extract user from req context
	// user attribute has been set prior in authMiddleware (auth.service.go)
	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	// get current timestamp in milliseconds
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	// insert message object into database
	result, _ := collection().InsertOne(ctx.Background(), bson.M{
		"authorId": user.UserID,
		"created":  tUnixMilli,
		"content":  message.Content,
	})

	// return created message object
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		created := getMessageById(oid.Hex())
		json.NewEncoder(w).Encode(created)
	}
}

/**
 * This function deletes a message
 * Requires messageId and author
 */
func deleteMessage(w http.ResponseWriter, req *http.Request) {
	// convert id (string) to objectId, by extracting id parameter from params
	messageId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid MessageID")
		return
	}

	// extract user from req context
	// user attribute has been set prior in authMiddleware (auth.service.go)
	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	// delete message object with id messageId and authorId UserID
	result, _ := collection().DeleteOne(ctx.Background(), bson.M{"_id": messageId, "authorId": user.UserID})

	// no message has been deleted => message by author with id doesnt exist
	// => print error
	if result.DeletedCount == 0 {
		printError(w, "Message not found")
		return
	}

	// print success
	json.NewEncoder(w).Encode(Exception{"Message deleted"})
}

func patchMessage(w http.ResponseWriter, req *http.Request) {
	// convert id (string) to objectId, by extracting id parameter from params
	messageId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid MessageID")
		return
	}

	// decode body into new message object
	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	// create a new validator
	validate := validator.New()
	// validate struct (validation defined in message/types.go (validate:"..."))
	// for additional information, check out: https://github.com/go-playground/validator
	err = validate.Struct(message)

	if err != nil {
		printError(w, "Bitte gib eine Nachricht ein")
		return
	}

	// extract user from req context
	// user attribute has been set prior in authMiddleware (auth.service.go)
	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	// get current timestamp in milliseconds
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	// update message object with id messageId and authorId UserID
	// overwrite old content and set updated to tUnixMilli
	result, _ := collection().UpdateOne(ctx.Background(), bson.M{"_id": messageId, "authorId": user.UserID}, bson.M{"$set": bson.M{"content": message.Content, "updated": tUnixMilli}})

	// no message has been updated => message by author with id doesnt exist
	// => print error
	if result.MatchedCount == 0 {
		printError(w, "Message not found "+fmt.Sprint(tUnixMilli))
		return
	}

	// return updated message
	updated := getMessageById(messageId.Hex())
	json.NewEncoder(w).Encode(updated)
}
