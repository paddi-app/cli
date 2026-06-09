package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Device-flow polling states (RFC 8628).
var (
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrSlowDown             = errors.New("slow down")
	ErrExpiredToken         = errors.New("device code expired")
)

func (c *Client) DeviceCode(ctx context.Context) (*DeviceCode, error) {
	var out DeviceCode
	if _, err := c.do(ctx, http.MethodPost, "/auth/device/code", nil, nil, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeviceToken(ctx context.Context, deviceCode string) (*Tokens, error) {
	var out Tokens
	body := map[string]string{"device_code": deviceCode}
	if _, err := c.do(ctx, http.MethodPost, "/auth/device/token", nil, body, &out, false); err != nil {
		var apiErr *Error
		if errors.As(err, &apiErr) {
			switch strings.ToLower(apiErr.Code) {
			case "authorization_pending":
				return nil, ErrAuthorizationPending
			case "slow_down":
				return nil, ErrSlowDown
			case "expired_token":
				return nil, ErrExpiredToken
			}
		}
		return nil, err
	}
	return &out, nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	var out Tokens
	body := map[string]string{"refresh_token": refreshToken}
	if _, err := c.do(ctx, http.MethodPost, "/auth/refresh-token", nil, body, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	body := map[string]string{"refresh_token": refreshToken}
	_, err := c.do(ctx, http.MethodPost, "/auth/logout", nil, body, nil, false)
	return err
}

func (c *Client) CurrentUser(ctx context.Context) (*User, json.RawMessage, error) {
	var out User
	raw, err := c.get(ctx, "/user", nil, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, raw, nil
}
