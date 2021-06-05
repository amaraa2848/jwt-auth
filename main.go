package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type RestResponseWithBody struct {
	IsOk bool `json:"is_ok"`
	Msg  string `json:"msg"`
	Data struct{
		Access_token string `json:"access_token"`
		Ca_token string `json:"ca_token"`
		Is_valid bool `json:"is_valid"`
	} `json:"data"`
}

type RestResponse struct{
	IsOk bool `json:"is_ok"`
	Msg  string `json:"msg"`
}


func main() {

	r := mux.NewRouter()
	r.HandleFunc("/v1/auth", GetTokenHandler).Methods(http.MethodPost)               // login
	r.HandleFunc("/v1/auth/validate", ValidateTokenHandler).Methods(http.MethodPost) // check token
	r.HandleFunc("/v1/user", CreateUserHandler).Methods(http.MethodPost)             // create user
	r.HandleFunc("/v1/user", UpdateUserHandler).Methods(http.MethodPut)              // update user info
	r.HandleFunc("/v1/user", DeleteUserHandler).Methods(http.MethodDelete)           // delete user
	http.ListenAndServe(":8080", r)
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	
	restResponse := RestResponse{}

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := new(User)
	err := json.Unmarshal(body.Bytes(), user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = fmt.Sprintf("Parsing user info failed: %s", err.Error())
		resJson,_ := json.Marshal(restResponse)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resJson)
		return
	}

	result, err := authenticateUser(ctx, user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error while authenticating user. Try again."
		resJson,_ := json.Marshal(restResponse)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)

		return
	}

	if result {
		token, err := generateToken(user.Email)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		user.JWT_token = token
		err = putUser(ctx, user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error while save user info. Try again."))
		}

		w.Write([]byte("Token created successfully!\n" + token))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User authentication failed!"))
	}

}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)
	result, err := validateToken(body.String())
	if err != nil {
		w.Write([]byte("Error occured while validating token! " + err.Error()))
	}
	if result {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}

}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Error parsing json body."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !user.checkField() {
		w.Write([]byte("Missing filed\n"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(&user)
	putUser(ctx, &user)
	if err != nil {
		w.Write([]byte("Insert failed.\n."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte("User created successfully.\n"))

}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Error parsing json body."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	old_user, err := getUser(ctx, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	if !user.checkField() {
		w.Write([]byte("Missing filed."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.CA_token = old_user.CA_token
	user.JWT_token = old_user.JWT_token

	putUser(ctx, &user)

}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Error parsing json body."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deleteUser(ctx, user.Email)
	if err != nil {
		w.Write([]byte("Error delete user."))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Success delete user."))
}
