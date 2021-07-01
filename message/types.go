package message

import "go.mongodb.org/mongo-driver/bson/primitive"

type Message struct {
	MessageID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AuthorID  primitive.ObjectID `json:"authorId,omitempty" bson:"authorId,omitempty"`
	Created   int64              `json:"created,omitempty" bson:"created,omitempty"`
	Updated   int64              `json:"updated,omitempty" bson:"updated,omitempty"`
	Content   string             `json:"content,omitempty" bson:"content,omitempty" validate:"required,gt=0"`
}

type Exception struct {
	Message string `json:"message"`
}
