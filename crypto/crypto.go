package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordManager struct {
	cost   int
	secret string
}

func NewPasswordManager(secret string) *PasswordManager {
	return &PasswordManager{
		cost:   bcrypt.DefaultCost,
		secret: secret,
	}
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	passwordWithSecret := fmt.Sprintf("%s%s", password, pm.secret)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSecret), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (pm *PasswordManager) ComparePassword(hashedPassword, password string) bool {
	passwordWithSecret := fmt.Sprintf("%s%s", password, pm.secret)

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwordWithSecret))
	return err == nil
}
