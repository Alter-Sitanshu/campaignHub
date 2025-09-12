package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrInvalidKeySize = fmt.Errorf("required key size : %d", chacha20poly1305.KeySize)
)

// Implements a
type PASETOMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// Creates a new paseto maker instance
func NewPASETOMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) != chacha20poly1305.KeySize {
		return nil, ErrInvalidKeySize
	}
	// return the struct
	return &PASETOMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(secretKey),
	}, nil
}

// verifies a paseto token
func (p *PASETOMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := p.paseto.Decrypt(token, p.symmetricKey, payload, nil)
	if err != nil {
		log.Printf("error decrypting token: %v\n", err.Error())
		return nil, err
	}
	if payload.IssuedAt.After(payload.ExpiredAt) {
		return nil, ErrTokenExpired
	}
	return payload, nil
}

// creates a paseto token based on the claims
func (p *PASETOMaker) CreateToken(iss, aud, email string, dur time.Duration) (string, error) {
	payload, err := NewPayload(iss, aud, email, dur)
	if err != nil {
		log.Printf("error creating PASETO payload\n")
		return "", err
	}
	// Here i used Encrypt function rather than Sign to use the v2.local PASETO
	// If you want, you can use paseto.Sign method which creates v2.public token
	token, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	if err != nil {
		log.Printf("error :%v\n", err.Error())
		return "", err
	}

	return token, nil
}

// Key Note: The Auth service signs/encrypts the payload and sends it to the client
// The client presents the token with each request and the Backend API service just verifies
// Whether i trust this token or not ?

// In case of v2.public API service has the public key for our auth service
// In case of v2.local API has the same symmetric key for out auth service(same trust level)
