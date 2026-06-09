package credentials

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	service         = "paddi-cli"
	accessTokenKey  = "access_token"
	refreshTokenKey = "refresh_token"
)

var ErrNotLoggedIn = errors.New("not logged in: run `paddi auth login`")

// FromEnv reports whether the access token comes from PADDI_TOKEN.
func FromEnv() bool { return os.Getenv("PADDI_TOKEN") != "" }

// AccessToken returns the access token, preferring PADDI_TOKEN over the keyring.
func AccessToken() (string, error) {
	if t := os.Getenv("PADDI_TOKEN"); t != "" {
		return t, nil
	}
	t, err := keyring.Get(service, accessTokenKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotLoggedIn
	}
	return t, err
}

func RefreshToken() (string, error) {
	t, err := keyring.Get(service, refreshTokenKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotLoggedIn
	}
	return t, err
}

func Store(accessToken, refreshToken string) error {
	if err := keyring.Set(service, accessTokenKey, accessToken); err != nil {
		return err
	}
	return keyring.Set(service, refreshTokenKey, refreshToken)
}

func Clear() error {
	for _, key := range []string{accessTokenKey, refreshTokenKey} {
		if err := keyring.Delete(service, key); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return err
		}
	}
	return nil
}
