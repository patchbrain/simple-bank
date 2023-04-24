package token

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type Payload struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	IssueAt  time.Time `json:"issue_at"`
	ExpireAt time.Time `json:"expire_at"`
}

var (
	ErrExpireToken  = errors.New("expired token")
	ErrInvalidToken = errors.New("invalid token")
)

func (p Payload) Valid() error {
	if time.Now().After(p.ExpireAt) {
		return ErrExpireToken
	}

	return nil
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Payload{
		Id:       id,
		Username: username,
		IssueAt:  time.Now(),
		ExpireAt: time.Now().Add(duration),
	}, nil
}
