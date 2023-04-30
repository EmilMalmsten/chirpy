package auth

import (
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrDoesNotMatch = errors.New("does not match")

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	storedHash := base64.StdEncoding.EncodeToString(hash)
	return storedHash, nil
}

func CheckPasswordHash(password, storedHash string) error {
	hash, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return fmt.Errorf("failed to decode stored hash: %s", err)
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		// password does not match
		return ErrDoesNotMatch
	} else {
		// password matches
		return nil
	}
}
