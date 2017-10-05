package model

import (
	"context"
	"fmt"
	"net/url"
)

// Authentication request
type authRequest struct {
	Type     string `json:"grant_type"`
	ClientID string `json:"client_id"`
	Secret   string `json:"client_secret"`
}

// Authentication Reply
type authReply struct {
	Token   string `json:"access_token"`
	Expires int    `json:"expires_in"`
	Type    string `json:"token_type"`
	Scope   string `json:"scope"`
}

// Login into clearpass with the provided credentials, return token.
func (c *clearpass) Login(ctx context.Context, address, clientID, secret string) (string, error) {
	baseURL := fmt.Sprintf("https://%s/api/", url.PathEscape(address))
	fullURL := baseURL + "oauth"
	req := authRequest{
		Type:     "client_credentials",
		ClientID: clientID,
		Secret:   secret,
	}
	rep := authReply{}
	if err := rest(ctx, POST, fullURL, "", nil, req, &rep, c.unsafe); err != nil {
		return "", err
	}
	c.url = baseURL
	c.token = rep.Token
	return c.token, nil
}

// Validate the token is still useful
func (c *clearpass) Validate(ctx context.Context, address, clientID, token string) error {
	baseURL := fmt.Sprintf("https://%s/api/", url.PathEscape(address))
	fullURL := baseURL + "api-client/" + url.PathEscape(clientID)
	rep := Reply{}
	if err := rest(ctx, GET, fullURL, token, nil, nil, &rep, c.unsafe); err != nil {
		return err
	}
	c.url = baseURL
	c.token = token
	return nil
}
