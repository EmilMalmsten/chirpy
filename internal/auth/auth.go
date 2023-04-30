package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func CreateToken(userId int, secretKey []byte, expiresInSeconds int) (string, error) {

	// Determine the expiration time
	var expTime time.Time
	if expiresInSeconds <= 0 {
		expiresInSeconds = 24 * 60 * 60 // default: 24 hours
	} else if expiresInSeconds > 24*60*60 {
		expiresInSeconds = 24 * 60 * 60 // max: 24 hours
	}
	expTime = time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	jwtExpTime := jwt.NewNumericDate(expTime)

	type Claims struct {
		jwt.RegisteredClaims
	}

	// Create claims with multiple fields populated
	claims := Claims{
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwtExpTime,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chirpy",
			Subject:   strconv.Itoa(userId),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token with key: %s", err)
	}

	return ss, nil

}
