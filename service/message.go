package service

import (
	"context"
	"errors"
	"gofeed-go/helper"
	"gofeed-go/persistence"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageService struct {
	p *persistence.MessagePersistor
}

func NewMessageService(p *persistence.MessagePersistor) *MessageService {
	return &MessageService{p}
}

var (
	ErrNotAuthor       = errors.New("Dieser Beitrag ist nicht von dir und kann deshalb nicht gelöscht werden.")
	ErrInvalidObjectID = errors.New("invalid ObjectID")
)

func (s *MessageService) GetMessages(ctx context.Context, limit *int64, skip *int64) (*[]persistence.Message, error) {
	opt := options.Find()

	if limit != nil {
		opt.SetLimit(*limit)
	}
	if skip != nil {
		opt.SetSkip(*skip)
	}

	return s.p.Find(ctx, bson.M{}, opt)
}

func (s *MessageService) GetMessageById(ctx context.Context, id string) (*persistence.Message, error) {
	return s.p.FindById(ctx, id)
}

func (s *MessageService) UpdateMessage(ctx context.Context, id string, author string, message persistence.Message) (*persistence.Message, error) {

	if !s.p.IsAuthor(ctx, id, author) {
		return nil, ErrNotAuthor
	}

	return s.p.UpdateById(ctx, id, message)
}

func (s *MessageService) CreateMessage(ctx context.Context, message persistence.Message) (*persistence.Message, error) {
	current := helper.GetCurrentTimeMillies()
	message.Created = current
	message.Updated = current

	return s.p.Create(ctx, message)
}

func (s *MessageService) DeleteMessage(ctx context.Context, id string, author string) (bool, error) {
	if !s.p.IsAuthor(ctx, id, author) {
		return false, ErrNotAuthor
	}

	return s.p.Delete(ctx, id)
}
