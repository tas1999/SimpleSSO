package cryptos

import (
	"crypto"
	"encoding/base64"

	"github.com/dgrijalva/jwt-go/v4"
)

type Secret struct {
	PasswordSecret  string `mapstructure:"password_secret"`
	PasswordHashAlg string `mapstructure:"password_hash_alg"`
	JwtSecret       string `mapstructure:"jwt_secret"`
	JwtAlg          string `mapstructure:"jwt_alg"`
}

func (a *Secret) GetHash(password string) (string, error) {
	hash := crypto.SHA256.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(hash.Sum([]byte(a.PasswordSecret))), nil
}
func (a *Secret) GetJwt(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.JwtSecret))
}
