package main

import (
	"context"
	"encoding/json"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type User struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	JWT_token    string `json:"jwt_token"`
	CA_token     string `json:"ca_token"`
	Access_level string `json:"access_level"`
}

func (user *User) checkField() bool {
	if len(user.Email) < 5 || len(user.Username) == 0 || len(user.Access_level) == 0 {
		return false
	}
	return true
}

type EtcdConfig struct {
	Endpoints []string
	Username  string
	Password  string
}

func initEtcdClient(config *EtcdConfig) *clientv3.Client {

	cli, err := clientv3.New(clientv3.Config{Endpoints: config.Endpoints})
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return cli
}

func getUser(ctx context.Context, email string) (*User, error) {
	cli := initEtcdClient(&EtcdConfig{Endpoints: []string{"localhost:2379"}})
	defer cli.Close()
	get_result, err := cli.Get(ctx, email)

	if err != nil {
		return nil, err
	}

	if get_result.Count == 0 {
		return nil, nil
	}

	user := new(User)
	fmt.Println(string(get_result.Kvs[0].Value))
	json.Unmarshal(get_result.Kvs[0].Value, user)

	return user, nil
}

func saveUser(ctx context.Context, user *User) error {
	cli := initEtcdClient(&EtcdConfig{Endpoints: []string{"localhost:2379"}})
	defer cli.Close()

	res, err := getUser(ctx, user.Email)
	if err != nil {
		return err
	}

	if res == nil {
		fmt.Println("Generate token for new user")
		newToken, err := generateToken(user.Email)
		if err != nil {
			return err
		}
		user.JWT_token = newToken
	}

	user_byte, err := json.Marshal(user)
	if err != nil {
		return err
	}

	put_result, err := cli.Put(ctx, user.Email, string(user_byte))
	if err != nil {
		return err
	}

	fmt.Println(put_result)

	return nil
}

func deleteUser(ctx context.Context, email string) error {
	cli := initEtcdClient(&EtcdConfig{Endpoints: []string{"localhost:2379"}})
	defer cli.Close()

	del_result, err := cli.Delete(ctx, email)
	if err != nil {
		return err
	}
	fmt.Println(del_result)
	return nil
}

func authenticateUser(ctx context.Context, user *User) (bool, error) {
	result, err := getUser(ctx, user.Email)
	if err != nil {
		return false, err
	}
	if result.Password != user.Password {
		return false, nil
	}
	return true, nil
}
