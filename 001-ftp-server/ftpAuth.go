package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Account ...
type Account struct {
	User string
	Pass string
	Dir  string
}

// CreateAccountListFromFile ...
func CreateAccountListFromFile(accountFile string) ([]Account, error) {
	file, err := os.OpenFile(accountFile, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return make([]Account, 0), err
	}
	reader := bufio.NewReader(file)
	accounts := make([]Account, 0)
	for user, err := reader.ReadString('\n'); err == nil; user, err = reader.ReadString('\n') {
		strs := strings.Fields(user)
		user = strs[0]
		pass := strs[1]
		dir := strs[2]
		account := Account{user, pass, dir}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Authenticate ...
func Authenticate(user string, pass string, accounts []Account) (string, error) {
	for _, v := range accounts {
		if v.User == user && v.Pass == pass {
			return v.Dir, nil
		}
	}
	return "", fmt.Errorf("user %v cannot login", user)
}
