package auth

import (
	"testing"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
)

func randomCreds(payload *Payload) {
	payload.Issuer = internals.RandString(5)
	payload.Audience = internals.RandString(5)
}

func TestInvalidKeySize(t *testing.T) {
	t.Run("invalid key size", func(t *testing.T) {
		secretKey := internals.RandString(30)
		_, err := NewJWTMaker(secretKey)
		if err == nil {
			t.Fail()
		}
	})
}

func TestJWT(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(time.Minute*5).Unix(), 0),
	}
	randomCreds(payload)
	t.Run("generating a valid JWT Token", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		if err != nil {
			t.Fail()
		}

		token, err := maker.CreateToken(payload.Issuer, payload.Audience, payload.Sub, time.Minute*5)
		if err != nil {
			t.Fail()
		}

		verifiedPayload, err := maker.VerifyToken(token)
		if err != nil {
			t.Fail()
		}
		if payload.IssuedAt != verifiedPayload.IssuedAt ||
			payload.ExpiredAt != verifiedPayload.ExpiredAt ||
			payload.Sub != verifiedPayload.Sub || payload.Issuer != verifiedPayload.Issuer ||
			payload.Audience != verifiedPayload.Audience {
			t.Fail()
		}

		if time.Now().After(verifiedPayload.ExpiredAt) {
			t.Fail()
		}

	})
}

func TestExpiredToken(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(time.Minute*5).Unix(), 0),
	}
	randomCreds(payload)
	t.Run("generating an expired JWT Token", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		if err != nil {
			t.Fail()
		}

		token, err := maker.CreateToken(payload.Issuer, payload.Audience, payload.Sub, -time.Minute*5)
		if err != nil {
			t.Fail()
		}

		_, err = maker.VerifyToken(token)
		if err == nil {
			t.Fail()
		}
		if err != ErrTokenExpired {
			t.Fail()
		}

	})
}

func TestInvalidToken(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(time.Minute*5).Unix(), 0),
	}
	t.Run("generating an invalid JWT Token", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		if err != nil {
			t.Fail()
		}

		_, err = maker.CreateToken(payload.Issuer, payload.Audience, payload.Sub, time.Minute*5)
		if err == nil {
			t.Fail()
		}

	})
}
