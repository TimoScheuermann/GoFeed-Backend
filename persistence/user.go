package persistence

import (
	"context"
	"gofeed-go/helper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserPersistor struct {
	c *mongo.Collection
}

type User struct {
	UserID      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" gofeed:"remUpdate,remInsert"`
	ProviderID  string             `json:"providerId" bson:"providerId" gofeed:"remUpdate"`
	Provider    string             `json:"provider" bson:"provider" gofeed:"remUpdate"`
	Name        string             `json:"name" bson:"name"`
	Avatar      string             `json:"avatar" bson:"avatar"`
	Group       string             `json:"group" bson:"group"`
	MemberSince int64              `json:"member_since" bson:"member_since" gofeed:"remUpdate"`
	LastLogin   int64              `json:"last_login" bson:"last_login"`
}

func NewUserPersistor(c *mongo.Collection) *UserPersistor {
	return &UserPersistor{c}
}

func (p *UserPersistor) FindById(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, ErrInvalidObjectID
	}

	res := p.c.FindOne(ctx, bson.M{"_id": oid})

	if res.Err() != nil {
		return nil, res.Err()
	}

	var user User
	err = res.Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *UserPersistor) FindByProvider(ctx context.Context, provider string, providerId string) (*User, error) {
	res := p.c.FindOne(ctx, bson.M{"provider": provider, "providerId": providerId})

	var user User
	err := res.Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *UserPersistor) Create(ctx context.Context, user User) (*User, error) {
	userCleaned := helper.CleanCreateBody(user)

	res, err := p.c.InsertOne(ctx, userCleaned)

	if err != nil {
		return nil, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return p.FindById(ctx, oid.Hex())
	}

	return nil, ErrInsertError

}
func (p *UserPersistor) Update(ctx context.Context, id string, update User) (*User, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	userCleaned := helper.CleanUpdateBody(update)

	res := p.c.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": userCleaned}, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if res.Err() != nil {
		return nil, res.Err()
	}

	var user User
	err = res.Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
