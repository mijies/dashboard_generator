package account

import (
	"log"
	"github.com/mijies/dashboard_builder/utils"
)

const DEFAULT_ACCOUNT_TXT = "samples/dashboard.txt"

type UserAccount struct {
	Name		string
	Login_name	string
}

func NewUserAccount(name string, login_name string) *UserAccount {
	if name == "" || login_name == "" {
		// return get_account_from_path(DEFAULT_ACCOUNT_TXT)
		return get_account_from_dialogbox()
	}
	return &UserAccount{
		Name: 		name,
		Login_name: login_name,
	}
}

func get_account_from_path(path string) *UserAccount {
	lines := utils.ReadLineSlice(path)
	if len(lines) != 2 {
		log.Fatal(DEFAULT_ACCOUNT_TXT + " must have user and login user names in 2 lines")
	}
	return &UserAccount{
		Name: 		lines[0],
		Login_name: lines[1],
	}
}
