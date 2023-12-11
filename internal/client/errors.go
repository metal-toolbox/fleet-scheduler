package client

import "github.com/pkg/errors"

var (
	ErrSsClientIsNil = errors.New("Serverservice client wasnt initialized")
	ErrCoClientIsNil = errors.New("Conditionorc client wasnt initialized")
	ErrNoTokenInRing = errors.New("secret not found in keyring")
	ErrAuth          = errors.New("authentication error")
	ErrNilConfig     = errors.New("configuration was nil")
)