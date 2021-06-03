package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Result struct {
	Token string `json:"access_token"`
	Date  string `json:"issued_at"`
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/v1/auth", GetTokenHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth", ValidateTokenHandler).Methods(http.MethodGet)
	http.ListenAndServe(":8080", r)
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is validate token"))
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	body := new(bytes.Buffer)

	io.Copy(body, r.Body)

	user := new(User)
	err := json.Unmarshal([]byte(body.String()), user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Parsing user info falied"))
		return
	}

	if result, _ := AuthenticateUser(user); result {
		token, err := generate(user.Username)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		retResult := new(Result)
		retResult.Date = time.Now().String()
		retResult.Token = token
		retJson, err := json.Marshal(retResult)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Building result body failed. Try again."))
		}

		w.Write([]byte(retJson))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Authentication failed!"))
	}

}
