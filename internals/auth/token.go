package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Payload struct {
	Id        uuid.UUID `json:"id"`
	Sub       string    `json:"sub"`
	IssuedAt  time.Time `json:"iat"`
	ExpiredAt time.Time `json:"exp"`
	Issuer    string    `json:"iss,omitempty"`
	Audience  string    `json:"aud,omitempty"`
}

func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.ExpiredAt), nil
}
func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}
func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}
func (payload *Payload) GetIssuer() (string, error) {
	return payload.Issuer, nil
}
func (payload *Payload) GetSubject() (string, error) {
	return payload.Sub, nil
}
func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{payload.Audience}, nil
}

func NewPayload(iss, aud, email string, dur time.Duration) (*Payload, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Printf("error making issued: %v\n", err.Error())
		return nil, err
	}
	if iss == "" || aud == "" {
		return nil, ErrMissingRequired
	}
	return &Payload{
		Id:        id,
		Sub:       email,
		Issuer:    iss,
		Audience:  aud,
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(dur).Unix(), 0),
	}, nil
}

type TokenMaker interface {
	VerifyToken(token string) (*Payload, error)
	CreateToken(iss, aud, email string, dur time.Duration) (string, error)
}
