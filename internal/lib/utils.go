package lib

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type StringAsBool bool

func (sb *StringAsBool) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "true", `"true"`:
		*sb = true
	case "false", `"false"`:
		*sb = false
	default:
		return errors.New("invalid value for boolean")
	}
	return nil
}

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

func RandomString(length uint) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
