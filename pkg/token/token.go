package token

import "time"

// Maker provides the all functions to create and verify any token.
type Maker interface {
	CreateToken(username string, duration time.Duration) (string, *Payload, error)
	VerfifyToken(token string) (*Payload, error)
}
