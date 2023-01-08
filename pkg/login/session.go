package gisha

import (
	"os/user"

	"github.com/zalando/go-keyring"
)

func ListSessions(appID string) (string, error) {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return keyring.Get(appID, currentUser.Username)
}

func CleanSession(appID string) error {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	// Delete the refresh token from the keyring
	return keyring.Delete(appID, currentUser.Username)
}
