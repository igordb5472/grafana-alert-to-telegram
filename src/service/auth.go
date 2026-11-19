package service

import (
	"fmt"
	"os"
)

var auth_login, auth_password string

func InjectAuthCreds() error {
	var ok bool

	auth_login, ok = os.LookupEnv("AUTH_LOGIN")
	if !ok {
		return fmt.Errorf("LoadAuthCreds: failed to load AUTH_LOGIN from env")
	}

	auth_password, ok = os.LookupEnv("AUTH_PASSWORD")
	if !ok {
		return fmt.Errorf("LoadAuthCreds: failed to load AUTH_PASSWORD from env")
	}

	return nil
}

func AuthOK(username, password string) bool {
	return username == auth_login && password == auth_password
}
