package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var signingKey = []byte("bnsoft")

func generateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"iat":   time.Now(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		println(err.Error())
		return "", err
	}
	return tokenString, nil
}

func validateToken(tokenString string) (bool, error) {
	// check token and get email from claims
	email, err := parseToken(tokenString)
	if err != nil {
		return false, err
	}
	
	// get user data by email and compare token
	ctx := context.Background()
	user, err := getUser(ctx, email)
	if err != nil {
		return false, err
	}

	if user.JWT_token != tokenString {
		return false, nil
	}
	
	return true, nil
}

func parseToken(tokenString string) (string, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return signingKey, nil
	})
	if err != nil {
		return "",err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email := claims["email"].(string)
		return email, nil
	} else {
		return "",errors.New("Invalid token")
	}
}