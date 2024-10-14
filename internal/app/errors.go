package app

import "github.com/pkg/errors"

var (
	ErrFdbClientIsNil = errors.New("FleetDB client wasnt initialized")
	ErrCoClientIsNil  = errors.New("Conditionorc client wasnt initialized")
	ErrNoTokenInRing  = errors.New("secret not found in keyring")
	ErrAuth           = errors.New("authentication error")
	ErrNilConfig      = errors.New("configuration was nil")
	ErrInvalidConfig  = errors.New("configuration is invalid")
	ErrConfig         = errors.New("configuration error")
)
