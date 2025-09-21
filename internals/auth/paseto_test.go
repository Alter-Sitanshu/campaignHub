package auth

import (
	"testing"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
)

func TestPasetoInvalidKeySize(t *testing.T) {
	t.Run("invalid key size", func(t *testing.T) {
		secretKey := internals.RandString(30)
		_, err := NewPASETOMaker([]byte(secretKey))
		if err == nil {
			t.Fail()
		}
	})
}

func TestPaseto(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Issuer:    "admin",
		Audience:  "admin",
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(time.Minute*5).Unix(), 0),
	}
	t.Run("generating a valid PASETO Token", func(t *testing.T) {
		maker, err := NewPASETOMaker([]byte(secretKey))
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
		if !payload.IssuedAt.Equal(verifiedPayload.IssuedAt) ||
			!payload.ExpiredAt.Equal(verifiedPayload.ExpiredAt) ||
			payload.Sub != verifiedPayload.Sub || payload.Issuer != verifiedPayload.Issuer ||
			payload.Audience != verifiedPayload.Audience {
			t.Fail()
		}

		if time.Now().After(verifiedPayload.ExpiredAt) {
			t.Fail()
		}

	})
}

func TestPasetoExpiredToken(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Issuer:    "admin",
		Audience:  "admin",
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(-time.Minute*5).Unix(), 0),
	}
	randomCreds(payload)
	t.Run("generating an expired PASETO Token", func(t *testing.T) {
		maker, err := NewPASETOMaker([]byte(secretKey))
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

func TestPasetoInvalidToken(t *testing.T) {
	secretKey := internals.RandString(32)
	payload := &Payload{
		Sub:       "useremail.com",
		IssuedAt:  time.Unix(time.Now().Unix(), 0),
		ExpiredAt: time.Unix(time.Now().Add(time.Minute*5).Unix(), 0),
	}
	t.Run("generating an invalid PASETO Token", func(t *testing.T) {
		maker, err := NewPASETOMaker([]byte(secretKey))
		if err != nil {
			t.Fail()
		}

		_, err = maker.CreateToken(payload.Issuer, payload.Audience, payload.Sub, time.Minute*5)
		if err == nil {
			t.Fail()
		}

	})
}
