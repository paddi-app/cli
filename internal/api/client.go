package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client talks HTTP to the Paddi backend.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	// Refresh, when set, is called once after a 401 to obtain a fresh
	// access token; the failed request is then retried.
	Refresh func(ctx context.Context) (string, error)
}

// Error is a typed backend error (framework.Exception).
type Error struct {
	Status int
	Code   string
	Meta   map[string]any
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s (HTTP %d)", e.Code, e.Status)
	}
	return fmt.Sprintf("HTTP %d", e.Status)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any, allowRefresh bool) (json.RawMessage, error) {
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}

	u := strings.TrimSuffix(c.BaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	refreshed := false
	for {
		req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}

		resp, err := c.httpClient().Do(req)
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized && allowRefresh && !refreshed && c.Refresh != nil {
			refreshed = true
			if token, err := c.Refresh(ctx); err == nil && token != "" {
				c.Token = token
				continue
			}
		}

		if resp.StatusCode >= 400 {
			return nil, parseError(resp.StatusCode, raw)
		}
		if out != nil && len(raw) > 0 {
			if err := json.Unmarshal(raw, out); err != nil {
				return nil, fmt.Errorf("decode response: %w", err)
			}
		}
		return raw, nil
	}
}

func parseError(status int, raw []byte) *Error {
	e := &Error{Status: status}
	var body struct {
		Code    string         `json:"code"`
		ErrCode string         `json:"error"`
		Meta    map[string]any `json:"meta"`
	}
	if json.Unmarshal(raw, &body) == nil {
		e.Code = body.Code
		if e.Code == "" {
			e.Code = body.ErrCode
		}
		e.Meta = body.Meta
	}
	return e
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, query, nil, out, true)
}

func (c *Client) post(ctx context.Context, path string, body, out any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, nil, body, out, true)
}

func (c *Client) patch(ctx context.Context, path string, body, out any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPatch, path, nil, body, out, true)
}
