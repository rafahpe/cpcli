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
func (c *clearpass) Login(ctx context.Context, address, clientID, secret string) (string, string, error) {
	req := authRequest{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": secret,
	}
	return c.auth(ctx, address, req)
}

// Validate the token is still useful
func (c *clearpass) Validate(ctx context.Context, address, clientID, token, refresh string) (string, string, error) {
	if refresh == "" {
		// No refresh, just check the current token is still valid.
		baseURL := apiURL(address)
		fullURL := baseURL + "api-client/" + url.PathEscape(clientID)
		rep := Reply{}
		if err := rest(ctx, GET, fullURL, token, nil, nil, &rep, c.unsafe); err != nil {
			return "", "", err
		}
		c.url, c.token, c.refresh = baseURL, token, ""
		return c.token, c.refresh, nil
	}
	req := authRequest{
		"grant_type":    "refresh_token",
		"client_id":     clientID,
		"refresh_token": refresh,
	}
	return c.auth(ctx, address, req)
}
