package main

import "errors"

func AuthenticateUser(user *User) (bool, error) {
	if user.Username == "amarmend" && user.Password == "bns2000" {
		return true, nil
	} else {
		return false, errors.New("user authentication failed!")
	}
}

func InsertToken() error {
	//TODO save generated token to database
	return nil
}

func ValidateToken(token string) (bool, error) {
	//TODO check if token exists and valid
	return true, nil
}
