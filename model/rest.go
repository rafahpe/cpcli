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

// HalLink is a link inside a struct
type halLink struct {
	Href string `json:"href"`
}

// HalLinks is returned by REST endpoints that return multiple objects
type halLinks struct {
	Self  halLink `json:"self"`
	First halLink `json:"first"`
	Last  halLink `json:"last"`
	Prev  halLink `json:"prev"`
	Next  halLink `json:"next"`
}

// WrappedItems wraps an array of Reply items
type wrappedItems struct {
	Items []RawReply `json:"items"`
}

// WrappedReply returned by endpoints that return multiple objects
type wrappedReply struct {
	Embedded wrappedItems `json:"_embedded"`
	Links    halLinks     `json:"_links"`
}

// Exhaust a channel, dismiss all replies
func Exhaust(replies Reply) {
	for replies.Next() {
	}
}

// Generic function to perform a REST request
func rest(ctx context.Context, method Method, url, token string, query map[string]string, request, reply interface{}, skipVerify bool) error {
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

// Follow a paginated stream
func follow(ctx context.Context, method Method, url, token string, query map[string]string, request interface{}, skipVerify bool) (Reply, error) {
	result := make(chan currentReply)
	waitme := make(chan error)
	go func() {
		defer close(result)
		var reply RawReply
		// Open the query and check for error.
		if err := rest(ctx, method, url, token, query, request, &reply, skipVerify); err != nil {
			waitme <- err
			return
		}
		// Release the caller
		close(waitme)
		// If it is not a wrapped reply, but a strange item, return it straight away
		var wReply wrappedReply
		if err := json.Unmarshal(reply, &wReply); err != nil {
			result <- currentReply{items: []RawReply{reply}}
			return
		}
		// If it is a wrappedReply, iterate on it
		for {
			result <- currentReply{items: wReply.Embedded.Items}
			// Run new request from link provided by HAL
			url := wReply.Links.Next.Href
			if url == "" || wReply.Links.Next.Href == wReply.Links.Self.Href {
				return
			}
			if err := rest(ctx, GET, url, token, nil, nil, &wReply, skipVerify); err != nil {
				result <- currentReply{lastError: err}
				return
			}
		}
	}()
	if err := <-waitme; err != nil {
		return nil, err
	}
	return &reply{stream: result}, nil
}
