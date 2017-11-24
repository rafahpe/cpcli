package model

import (
	"context"
	"net/url"
)

// Authentication request
type authRequest map[string]string

// Authentication Reply
type authReply struct {
	Token   string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Expires int    `json:"expires_in"`
	Type    string `json:"token_type"`
	Scope   string `json:"scope"`
}

// Perform an authentication request
func (c *clearpass) auth(ctx context.Context, address string, req authRequest) (string, string, error) {
	baseURL := apiURL(address)
	fullURL := baseURL + "oauth"
	rep := authReply{}
	if err := rest(ctx, POST, fullURL, "", nil, req, &rep, c.unsafe); err != nil {
		return "", "", err
	}
	c.url, c.token, c.refresh = baseURL, rep.Token, rep.Refresh
	return c.token, c.refresh, nil
}

// Login into clearpass with the provided credentials, return token.
func (c *clearpass) Login(ctx context.Context, address, clientID, secret, username, pass string) (string, string, error) {
	req := authRequest{
		"grant_type": "client_credentials",
		"client_id":  clientID,
	}
	if secret != "" {
		req["client_secret"] = secret
	}
	if username != "" && pass != "" {
		req["grant_type"] = "password"
		req["username"] = username
		req["password"] = pass
	}
	return c.auth(ctx, address, req)
}

// Validate the token is still useful
func (c *clearpass) Validate(ctx context.Context, address, clientID, secret, token, refresh string) (string, string, error) {
	// If there is arefresh token, try to refresh auth
	if refresh != "" {
		req := authRequest{
			"grant_type":    "refresh_token",
			"client_id":     clientID,
			"refresh_token": refresh,
		}
		if secret != "" {
			req["client_secret"] = secret
		}
		t, r, err := c.auth(ctx, address, req)
		if err == nil {
			return t, r, err
		}
	}
	// No refresh or it didn't succeed, just check the
	// current token is still valid.
	baseURL := apiURL(address)
	fullURL := baseURL + "api-client/" + url.PathEscape(clientID)
	rep := Reply{}
	if err := rest(ctx, GET, fullURL, token, nil, nil, &rep, c.unsafe); err != nil {
		return "", "", err
	}
	c.url, c.token, c.refresh = baseURL, token, ""
	return c.token, c.refresh, nil
}
