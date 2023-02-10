package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must have %v", chacha20poly1305.KeySize)
	}

	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}
func (p *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration) // creates a new payload.
	if err != nil {
		return "", nil, err
	}
	encrypted, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	return encrypted, payload, err
}

func (p *PasetoMaker) VerfifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := p.paseto.Decrypt(token, p.symmetricKey, payload, nil) // Decrypt by providing the token and the symmetricKey
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err = payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
