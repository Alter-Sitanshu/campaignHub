package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Payload struct {
	Id        string    `json:"id"`
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

func NewPayload(iss, aud, objID string, dur time.Duration) (*Payload, error) {
	id := uuid.NewString()
	if iss == "" || aud == "" {
		return nil, ErrMissingRequired
	}
	return &Payload{
		Id:        id,
		Sub:       objID,
		Issuer:    iss,
		Audience:  aud,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(dur),
	}, nil
}

type TokenMaker interface {
	VerifyToken(token string) (*Payload, error)
	CreateToken(iss, aud, objID string, dur time.Duration) (string, error)
}
