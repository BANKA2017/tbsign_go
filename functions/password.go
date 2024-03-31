package _function

import "golang.org/x/crypto/bcrypt"

func CreatePasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}

func VerifyPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
