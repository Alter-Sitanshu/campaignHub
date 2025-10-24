package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeylen = 32

// errors for JWT validator
var (
	ErrSigningMethod     = errors.New("unexpected signing method")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("expired token")
	ErrMissingRequired   = errors.New("missing aud or iss")
	ErrInvalidKeySizeJWT = fmt.Errorf("invalid key size: key must be: %d characters", minSecretKeylen)
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (TokenMaker, error) {
	// make sure the secret key length is valid
	if len(secretKey) < minSecretKeylen {
		log.Printf("invalid key size: key must be: %d characters", minSecretKeylen)
		return nil, ErrInvalidKeySizeJWT
	}
	// return the JWT Maker
	return &JWTMaker{secretKey: secretKey}, nil
}

func (j *JWTMaker) CreateToken(iss, aud, email string, dur time.Duration) (string, error) {
	payload, err := NewPayload(iss, aud, email, dur)
	if err != nil {
		log.Printf("error creating token: %v\n", err.Error())
		return "", err
	}
	// making the jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	// signing the jwt token
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		log.Printf("error signing token: %v\n", err.Error())
		return "", err
	}
	return tokenString, nil
}
func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{},
		func(t *jwt.Token) (any, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrInvalidToken
			}

			return []byte(j.secretKey), nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	// no error flags raised
	return payload, nil
}
