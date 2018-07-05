package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context/ctxhttp"
)

// ErrNotLoggedIn returned when the user is not logged into CPPM yet
const ErrNotLoggedIn = Error("Not authorized. Make sure you log in and your account has the proper privileges")

// Params for ClearPass request
type Params map[string]string

// Generic raw HTTP request
func rawRequest(ctx context.Context, client *http.Client, req *http.Request, body []byte) RestError {
	var bodyReader io.ReadCloser
	if body != nil {
		bodyReader = ioutil.NopCloser(bytes.NewReader(body))
	}
	detail := RestError{
		Method: Method(req.Method),
		URL:    req.URL.RequestURI(),
		Query:  req.URL.RawQuery,
		Header: req.Header,
		Body:   body,
	}
	req.Body = bodyReader
	resp, err := ctxhttp.Do(ctx, client, req)
	if resp != nil {
		detail.ReplyHeader = resp.Header
		if resp.Body != nil {
			defer resp.Body.Close()
		}
	}
	detail.Err = err
	detail.StatusCode = resp.StatusCode
	if err != nil {
		return detail
	}
	if resp.Body != nil {
		reply, err := ioutil.ReadAll(resp.Body)
		if detail.Err == nil {
			detail.Err = err
		}
		detail.Reply = reply
	}
	return detail
}

// Generic function to perform a REST request
func rest(ctx context.Context, client *http.Client, method Method, url, token string, query Params, request, reply interface{}) error {
	var jsonBody []byte
	var err error
	if request != nil {
		jsonBody, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(string(method), url, nil)
	if err != nil {
		return err
	}
	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if query != nil && len(query) > 0 {
		q := req.URL.Query()
		for key, val := range query {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/json")
	detail := rawRequest(ctx, client, req, jsonBody)
	if detail.Err != nil {
		return detail
	}
	if detail.StatusCode == 401 || detail.StatusCode == 403 {
		detail.Err = ErrNotLoggedIn
		return detail
	}
	if detail.StatusCode != 200 {
		detail.Err = fmt.Errorf("Error: REST Status %d", detail.StatusCode)
		return detail
	}
	if err := json.Unmarshal(detail.Reply, reply); err != nil {
		detail.Err = err
		return detail
	}
	return nil
}
