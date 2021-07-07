package persistence

import (
	"context"
	"errors"
	"gofeed-go/helper"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessagePersistor struct {
	c *mongo.Collection
}

type Message struct {
	MessageID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" gofeed:"remUpdate,remInsert"`
	AuthorID  primitive.ObjectID `json:"authorId,omitempty" bson:"authorId,omitempty" gofeed:"remUpdate"`
	Created   int64              `json:"created,omitempty" bson:"created,omitempty" gofeed:"remUpdate"`
	Updated   int64              `json:"updated,omitempty" bson:"updated,omitempty"`
	Content   string             `json:"content,omitempty" bson:"content,omitempty" validate:"required,gt=0"`
}

var (
	ErrInsertError     = errors.New("something ubiquitous happened")
	ErrInvalidObjectID = errors.New("invalid ObjectID")
	ErrNothingDeleted  = errors.New("nothing has been deleted")
	ErrMissingContent  = errors.New("Dein Beitrag ist zu kurz!")
	validate           = validator.New()
)

func NewMessagePersistor(c *mongo.Collection) *MessagePersistor {
	return &MessagePersistor{c}
}

func (p *MessagePersistor) FindById(ctx context.Context, id string) (*Message, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, ErrInvalidObjectID
	}

	res := p.c.FindOne(ctx, bson.M{"_id": oid})

	if res.Err() != nil {
		return nil, res.Err()
	}

	var message Message
	err = res.Decode(&message)

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (p *MessagePersistor) UpdateById(ctx context.Context, id string, update Message) (*Message, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, ErrInvalidObjectID
	}

	// validate struct (validation defined in message/types.go (validate:"..."))
	// for additional information, check out: https://github.com/go-playground/validator
	err = validate.Struct(update)
	if err != nil {
		return nil, ErrMissingContent
	}

	updateClaned := helper.CleanUpdateBody(update)

	res := p.c.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": updateClaned}, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if res.Err() != nil {
		return nil, res.Err()
	}

	var message Message
	err = res.Decode(&message)

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (p *MessagePersistor) Create(ctx context.Context, create Message) (*Message, error) {
	// validate struct (validation defined in message/types.go (validate:"..."))
	// for additional information, check out: https://github.com/go-playground/validator
	err := validate.Struct(create)
	if err != nil {
		return nil, ErrMissingContent
	}

	createCleaned := helper.CleanCreateBody(create)

	res, err := p.c.InsertOne(ctx, createCleaned)

	if err != nil {
		return nil, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return p.FindById(ctx, oid.Hex())
	}

	return nil, ErrInsertError
}

func (p *MessagePersistor) Find(ctx context.Context, filter bson.M, options *options.FindOptions) (*[]Message, error) {
	cursor, err := p.c.Find(ctx, filter, options)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	messages := []Message{}

	for cursor.Next(ctx) {
		var message Message
		err := cursor.Decode(&message)

		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return &messages, nil
}

func (p *MessagePersistor) Delete(ctx context.Context, id string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return false, err
	}

	res, err := p.c.DeleteOne(ctx, bson.M{"_id": oid})

	if err != nil {
		return false, err
	}

	if res.DeletedCount == 0 {
		return false, ErrNothingDeleted
	}

	return true, nil
}

func (p *MessagePersistor) IsAuthor(ctx context.Context, id string, author string) bool {
	mid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return false
	}

	aid, err := primitive.ObjectIDFromHex(author)

	if err != nil {
		return false
	}

	res := p.c.FindOne(ctx, bson.M{"_id": mid, "authorId": aid})

	return res.Err() == nil
}
