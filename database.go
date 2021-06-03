package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DBCredentials struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	Protocol string
}

var DBCred *DBCredentials

var (
	ErrInvalidDBCredential      = errors.New("Databse credential is invalid.")
	ErrDatabaseConnectionFailed = errors.New("Failed to connect database.")
	ErrQueryFailed              = errors.New("Failed to execute query on database. Try again")
	ErrUserAuthFailed           = errors.New("user authentication failed!")
)

func initDatabase(cred DBCredentials) error {

	if len(cred.Host) == 0 {
		cred.Host = "localhost"
	}

	if len(cred.Port) == 0 {
		cred.Port = "3306"
	}

	if len(cred.Username) == 0 {
		cred.Username = "root"
	}

	if len(cred.Protocol) == 0 {
		cred.Protocol = "tcp"
	}
	DBCred = &cred
	// check db credential by making a connection to actual database
	_, err := getDB()
	if err != nil {
		DBCred = nil
		return ErrInvalidDBCredential
	}

	return nil
}

func getDB() (*sql.DB, error) {

	if len(DBCred.DBName) == 0 {

	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%s)/%s", DBCred.Username, DBCred.Password, DBCred.Protocol, DBCred.Host, DBCred.Port, DBCred.DBName))

	if err != nil {
		return nil, err
	}
	return db, nil

}

func execQuery(query string) {
	db, err := getDB()
	if err != nil {

	}
	defer db.Close()
	println("QUERY: ", query)
	res, err := db.Exec(query)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	rows, _ := res.RowsAffected()
	fmt.Println("Succesfully affected:", rows)
}

func selectQuery(query string) (*sql.Rows, error) {
	db, _ := getDB()
	defer db.Close()
	println("QUERY: ", query)
	res, err := db.Query(query)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return nil, err
	}
	return res, nil
}

func AuthenticateUser(user *User) (bool, error) {
	q := fmt.Sprintf("SELCECT * FROM auth.TB_USER WHERE username = '%s' and password = %s", user.Username, user.Password)
	result, err := selectQuery(q)

	if err != nil {
		return false, ErrQueryFailed
	}
	if result.Next() {
		return true, nil
	} else {
		return false, ErrUserAuthFailed
	}
}

func InsertToken(username string) error {
	//TODO save generated token to database
	token, _ := generate(username)
	fmt.Println(token)
	return nil
}

func ValidateToken(token string) (bool, error) {
	//TODO check if token exists and valid
	return true, nil
}
