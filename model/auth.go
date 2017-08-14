package model

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rafahpe/cpcli/lib"
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

// Login into clearpass with the provided credentials
func (c *clearpass) Login(ip, clientID, secret string) (string, error) {
	baseURL := fmt.Sprintf("https://%s:443/api/", url.PathEscape(ip))
	fullURL := baseURL + "oauth"
	query := authRequest{
		Type:     "client_credentials",
		ClientID: clientID,
		Secret:   secret,
	}
	reply := &authReply{}
	ctx := context.Background()
	if err := post(ctx, fullURL, "", query, reply); err != nil {
		return "", err
	}
	c.url = baseURL
	c.token = reply.Token
	return c.token, nil
}

// Validate the token is still useful
func (c *clearpass) Validate(ip, clientID, token string) error {
	baseURL := fmt.Sprintf("https://%s:443/api/", url.PathEscape(ip))
	fullURL := baseURL + "api-client/" + url.PathEscape(clientID)
	reply := &lib.Reply{}
	ctx := context.Background()
	if err := get(ctx, fullURL, token, nil, reply); err != nil {
		return err
	}
	c.url = baseURL
	c.token = token
	return nil
}
