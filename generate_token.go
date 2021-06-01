package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

func generate(name string) (string, error) {
	signingKey := []byte("bnsoft")
	claims := jwt.MapClaims{
		"name": name,
		"iat":  time.Now(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		println(err.Error())
		return "", err
	}
	return tokenString, nil
}
