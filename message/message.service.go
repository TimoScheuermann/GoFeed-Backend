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

func collection() *mongo.Collection {
	return database.Database.Collection("message")
}

func printError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(Exception{message})
}

func getMessageById(messageId string) *Message {
	var message Message
	oid, err := primitive.ObjectIDFromHex(messageId)

	if err != nil {
		fmt.Printf("Invalid ObjectID %v\n", messageId)
		return nil
	}

	err = collection().FindOne(ctx.Background(), bson.M{"_id": oid}).Decode(&message)

	if err != nil {
		fmt.Printf("Message not found... (%v) %v \n", messageId, err)
		return nil
	}

	return &message

}

func getMessages(w http.ResponseWriter, req *http.Request) {
	options := options.Find()
	query := req.URL.Query()
	qLimit, qSkip := query.Get("limit"), query.Get("skip")

	if len(qLimit) > 0 {
		if limit, err := strconv.ParseInt(qLimit, 10, 64); err == nil {
			options.SetLimit(limit)
		}
	}

	if len(qSkip) > 0 {
		if skip, err := strconv.ParseInt(qSkip, 10, 64); err == nil {
			options.SetSkip(skip)
		}
	}

	cursor, err := collection().Find(ctx.Background(), bson.M{}, options)

	if err != nil {
		printError(w, err.Error())
		return
	}

	defer cursor.Close(ctx.Background())

	messages := []Message{}
	for cursor.Next(ctx.Background()) {
		var message Message
		cursor.Decode(&message)
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		printError(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(messages)
}

func getMessage(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	message := getMessageById(params["id"])

	if message == nil {
		printError(w, "Message not found")
		return
	}

	json.NewEncoder(w).Encode(message)
}

func postMessage(w http.ResponseWriter, req *http.Request) {
	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	validate := validator.New()
	err := validate.Struct(message)

	if err != nil {
		printError(w, "Bitte gib eine Nachricht ein")
		return
	}

	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	result, _ := collection().InsertOne(ctx.Background(), bson.M{
		"authorId": user.UserID,
		"created":  tUnixMilli,
		"content":  message.Content,
	})

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		created := getMessageById(oid.Hex())
		json.NewEncoder(w).Encode(created)
	}
}

func deleteMessage(w http.ResponseWriter, req *http.Request) {
	messageId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid MessageID")
		return
	}

	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	result, _ := collection().DeleteOne(ctx.Background(), bson.M{"_id": messageId, "authorId": user.UserID})

	if result.DeletedCount == 0 {
		printError(w, "Message not found")
		return
	}

	json.NewEncoder(w).Encode(Exception{"Message deleted"})
}

func patchMessage(w http.ResponseWriter, req *http.Request) {
	messageId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid MessageID")
		return
	}

	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	validate := validator.New()
	err = validate.Struct(message)

	if err != nil {
		printError(w, "Bitte gib eine Nachricht ein")
		return
	}

	var user auth.User
	mapstructure.Decode(context.Get(req, "user"), &user)

	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	result, _ := collection().UpdateOne(ctx.Background(), bson.M{"_id": messageId, "authorId": user.UserID}, bson.M{"$set": bson.M{"content": message.Content, "updated": tUnixMilli}})

	if result.MatchedCount == 0 {
		printError(w, "Message not found "+fmt.Sprint(tUnixMilli))
		return
	}

	updated := getMessageById(messageId.Hex())
	json.NewEncoder(w).Encode(updated)
}
