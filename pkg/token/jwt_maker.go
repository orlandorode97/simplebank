package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const minSecretSize = 12

// JWTMaker structs stores the secret key to sign the JWT.
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker returns an implementation of the Maker interface by providing the secretKey to sign the JWT
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretSize {
		return nil, fmt.Errorf("invalid secret key size: must be at least %v", minSecretSize)
	}

	return &JWTMaker{
		secretKey: secretKey,
	}, nil
}

// CreateToken creates a jwt token.
func (j *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration) // creates a new payload.
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signed, err := token.SignedString(j.secretKey)
	return signed, payload, err
}

// VerfifyToken verifies the JWT.
func (j *JWTMaker) VerfifyToken(token string) (*Payload, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(j.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		vErrr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(vErrr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}

		return nil, err
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok && payload.Valid() != nil {
		return nil, ErrInvalidToken
	}
	return nil, nil
}
