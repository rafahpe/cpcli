package model

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context/ctxhttp"
)

// ErrNotLoggedIn returned when the user is not logged into CPPM yet
const ErrNotLoggedIn = Error("Not authorized. Make sure you log in and your account has the proper privileges")

// Method is the HTTP method for the request
type Method string

// HTTP methods supported
const (
	GET    Method = "GET"
	POST          = "POST"
	PUT           = "PUT"
	UPDATE        = "UPDATE"
	DELETE        = "DELETE"
	PATCH         = "PATCH"
)

// Params for ClearPass request
type Params map[string]string

// Generic function to perform a REST request
func rest(ctx context.Context, method Method, url, token string, query Params, request, reply interface{}, skipVerify bool) error {
	details := RestError{Method: string(method), URL: url}
	var body io.Reader
	if request != nil {
		jsonBody, err := json.Marshal(request)
		if err != nil {
			return err
		}
		body = bytes.NewReader(jsonBody)
		details.Body = string(jsonBody)
	}
	req, err := http.NewRequest(string(method), url, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if query != nil && len(query) > 0 {
		q := req.URL.Query()
		for key, val := range query {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
		details.Query = req.URL.RawQuery
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/json")
	details.Header = req.Header
	details.Unsafe = skipVerify
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
		},
	}
	resp, err := ctxhttp.Do(ctx, client, req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		details.Err = err
		return details
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		details.Err = err
		return details
	}
	details.StatusCode = resp.StatusCode
	details.Reply = string(respBody)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		details.Err = ErrNotLoggedIn
		return details
	}
	if resp.StatusCode != 200 {
		details.Err = fmt.Errorf("Error: REST Status %s", resp.Status)
		return details
	}
	if err := json.Unmarshal(respBody, reply); err != nil {
		details.Err = err
		return details
	}
	return nil
}
