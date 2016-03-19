// provides connection encryption
package secure

import (
	"errors"
)

const (
	// fixed key size in byte
	KeySize = 32

	// rounds of iteration for pbkdf2
	KeyIteration = 4096
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)
