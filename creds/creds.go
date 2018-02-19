package creds

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLength = 4
	minPasswordLength = 8
)

// Ensures the username and password given to signup with meet our acceptance criteria
// Note: as of right now we only require usernames to be 4 characters long and passwords 8
//		 additionally username must be url safe
// The following characters are URL safe: ALPHA DIGIT "-" / "." / "_" / "~"
func ValidateSignupCredentials(username, password string) error {
	// Note: the error messages in this function are user facing

	if len(username) < minUsernameLength {
		return fmt.Errorf("Username must be at least %v characters long", minUsernameLength)
	}

	if len(password) < minPasswordLength {
		return fmt.Errorf("Password must be at least %v characters long", minPasswordLength)
	}

	for _, asciiVal := range []rune(username) {
		if !isURLSafe(asciiVal) {
			return fmt.Errorf("Usernames must only contain contain (a-z A-Z 0-9 - . _ ~) - found: %v", string(asciiVal))
		}
	}

	// Passwords must also be url safe for the sake of avoiding having strange characters in passwords
	for _, asciiVal := range []rune(password) {
		if !isURLSafe(asciiVal) {
			return fmt.Errorf("Passwords must only contain contain (a-z A-Z 0-9 - . _ ~)")
		}
	}

	return nil
}

// Determines that the given code point is URL safe
// The following characters are URL safe: ALPHA DIGIT "-" / "." / "_" / "~"
func isURLSafe(c rune) bool {
	// In otherwords !upperCase && !lowerCase && !number && !dash && !dot && !underscore && !tilda
	if (c > 90 || c < 65) && (c > 122 || c < 97) && (c < 48 || c > 57) && (c != 45) && (c != 46) && (c != 95) && (c != 126) {
		return false
	}

	return true
}

// Consumes plaintext password and hashes using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// User for authenticating login to compare password and hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
