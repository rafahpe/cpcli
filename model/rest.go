package model

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rafahpe/cpcli/lib"

	"golang.org/x/net/context/ctxhttp"
)

// Error type
type Error string

func (e Error) Error() string {
	return string(e)
}

// ErrNotLoggedIn returned when the user is not logged into CPPM yet
const ErrNotLoggedIn = Error("You must log into Clearpass to start")

// ErrPageTooSmall when paginated commands are givel a too small page size
const ErrPageTooSmall = Error("Page size is too small")

// HalLink is a link inside a struct
type halLink struct {
	Href string `json:"href"`
}

// HalLinks is returned by REST endpoints that manage multiple values
type halLinks struct {
	Self  halLink `json:"self"`
	First halLink `json:"first"`
	Last  halLink `json:"last"`
	Prev  halLink `json:"prev"`
	Next  halLink `json:"next"`
}

// WrappedItems wraps an array of items
type wrappedItems struct {
	Items []lib.Reply `json:"items"`
}

// WrappedReply returned by endpoints that provide multiples values
type wrappedReply struct {
	Embedded wrappedItems `json:"_embedded"`
	Links    halLinks     `json:"_links"`
}

// Generic function to perform a REST request
func rest(ctx context.Context, token string, request *http.Request, reply interface{}) error {
	if token != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	request.Header.Set("Accept", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := ctxhttp.Do(ctx, client, request)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return ErrNotLoggedIn
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: REST Status %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(reply)
}

func (w *wrappedReply) flush(output chan lib.Reply) {
	for _, item := range w.Embedded.Items {
		output <- item
	}
}

// Post Request can be a struct, reply must be pointer to struct
func post(ctx context.Context, url, token string, request, reply interface{}) error {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return rest(ctx, token, req, reply)
}

// Get Request can be an struct, reply must be pointer to struct
func get(ctx context.Context, url, token string, request map[string]string, reply interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if request != nil && len(request) > 0 {
		q := req.URL.Query()
		for key, val := range request {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	return rest(ctx, token, req, reply)
}

// Follow a paginated stream
func follow(ctx context.Context, url, token string, request map[string]string) chan lib.Reply {
	result := make(chan lib.Reply)
	go func(result chan lib.Reply) {
		defer close(result)
		logger := log.New(os.Stderr, "", 0)
		// First page, get as is
		reply := wrappedReply{}
		if err := get(ctx, url, token, request, &reply); err != nil {
			logger.Print(err)
			return
		}
		reply.flush(result)
		// Following pages, get from HAL
		for {
			// Run new request from link provided by HAL
			url := reply.Links.Next.Href
			if url == "" {
				return
			}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				logger.Print(err)
				return
			}
			// Get new reply
			reply = wrappedReply{}
			if err := rest(ctx, token, req, &reply); err != nil {
				logger.Print(err)
				return
			}
			reply.flush(result)
		}
	}(result)
	return result
}
