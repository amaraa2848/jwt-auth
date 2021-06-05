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

type RestResponse struct{
	IsOk bool `json:"is_ok"`
	Msg  string `json:"msg"`
}

type TokenValidationResponse struct{
	IsOk bool `json:"is_ok"`
	Msg  string `json:"msg"`
	IsValid bool `json:"is_valid"`
} 

type TokenGeneratingResponse struct{
	IsOk bool `json:"is_ok"`
	Msg  string `json:"msg"`
	Token string `json:"token"`
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
	tokenGeneratingResponse := TokenGeneratingResponse{}

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

	if !result {
		restResponse.IsOk = false
		restResponse.Msg = "User authentication failed!"
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resJson)
		return
	}
	token, err := generateToken(user.Email)

	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Token generation failed!"
		resJson,_ := json.Marshal(restResponse)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
		return
	}
	user.JWT_token = token
	err = saveUser(ctx, user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error while saving generated token. Try again."
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
		return
	}
	tokenGeneratingResponse.IsOk = true
	tokenGeneratingResponse.Msg = "Token created successfully."
	tokenGeneratingResponse.Token = token
	resJson,_ := json.Marshal(tokenGeneratingResponse)
	w.Write(resJson)
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	restResponse := RestResponse{}
	tokenValidationResponse := TokenValidationResponse{}

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)
	result, err := validateToken(body.String())
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = fmt.Sprintf("Error occured while validating token!\n %s", err.Error())
		resJson,_ := json.Marshal(restResponse)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
		return
	}
	if !result {
		tokenValidationResponse.IsOk = true
		tokenValidationResponse.Msg = "Token is invalid"
		tokenValidationResponse.IsValid = false
		resJson,_ := json.Marshal(tokenValidationResponse)
		w.Write(resJson)
		return
	} 
	
	tokenValidationResponse.IsOk = true
	tokenValidationResponse.Msg = "Token is valid"
	tokenValidationResponse.IsValid = true
	resJson,_ := json.Marshal(tokenValidationResponse)
	w.Write(resJson)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	restResponse := RestResponse{}

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = fmt.Sprintf("Error parsing json body!\n %s", err.Error())
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resJson)
		return
	}
	if !user.checkField() {
		restResponse.IsOk = false
		restResponse.Msg = "One or more required field if missing."
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resJson)
		return
	}
	fmt.Println(&user)
	err = saveUser(ctx, &user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Insert failed.\n" + err.Error()
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
		return
	}
	restResponse.IsOk = true
	restResponse.Msg = "User created successfully.\n"
	resJson,_ := json.Marshal(restResponse)
	w.Write(resJson)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	restResponse := RestResponse{}

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error parsing json body.\n" + err.Error()
		resJson,_ := json.Marshal(restResponse)
		w.Write(resJson)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	old_user, err := getUser(ctx, user.Email)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error while getting \n" + err.Error()
		resJson,_ := json.Marshal(restResponse)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
	}
	if !user.checkField() {
		restResponse.IsOk = false
		restResponse.Msg = "One or more required field if missing."
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resJson)
		return
	}
	user.CA_token = old_user.CA_token
	user.JWT_token = old_user.JWT_token

	saveUser(ctx, &user)
	err = saveUser(ctx, &user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error occured while updating user info."
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resJson)
		return
	}
	restResponse.IsOk = true
	restResponse.Msg = "User info updated successfully.\n"
	resJson,_ := json.Marshal(restResponse)
	w.Write(resJson)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	restResponse := RestResponse{}

	body := new(bytes.Buffer)
	io.Copy(body, r.Body)

	ctx := context.Background()

	user := User{}
	err := json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Error parsing json body."
		resJson,_ := json.Marshal(restResponse)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resJson)
		return
	}

	err = deleteUser(ctx, user.Email)
	if err != nil {
		restResponse.IsOk = false
		restResponse.Msg = "Failed to delete user."
		resJson,_ := json.Marshal(restResponse)

		w.Write([]byte(resJson))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	restResponse.IsOk = true
	restResponse.Msg = "Successfully deleted user."
	resJson,_ := json.Marshal(restResponse)
	w.Write(resJson)
}
