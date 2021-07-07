package service

import (
	"context"
	"gofeed-go/helper"
	"gofeed-go/persistence"

	"github.com/markbates/goth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	p *persistence.UserPersistor
}

type UserInfo struct {
	UserID primitive.ObjectID `json:"id"`
	Name   string             `json:"name"`
	Avatar string             `json:"avatar"`
}

func NewUserService(p *persistence.UserPersistor) *UserService {
	return &UserService{p}
}

func (s *UserService) UserSignedIn(ctx context.Context, gothUser goth.User) (*persistence.User, error) {
	millis := helper.GetCurrentTimeMillies()

	user, err := s.p.FindByProvider(ctx, gothUser.Provider, gothUser.UserID)

	if err != nil || user == nil {
		user, err = s.p.Create(ctx, persistence.User{
			ProviderID:  gothUser.UserID,
			Provider:    gothUser.Provider,
			Name:        gothUser.Name,
			Avatar:      gothUser.AvatarURL,
			Group:       "user",
			MemberSince: millis,
			LastLogin:   millis,
		})
	} else {
		user, err = s.p.Update(ctx, user.UserID.Hex(), persistence.User{
			Name:      gothUser.Name,
			Avatar:    gothUser.AvatarURL,
			LastLogin: millis,
		})
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, id string) (*UserInfo, error) {
	user, err := s.p.FindById(ctx, id)

	if err != nil {
		return nil, err
	}

	return &UserInfo{UserID: user.UserID, Name: user.Name, Avatar: user.Avatar}, nil
}
