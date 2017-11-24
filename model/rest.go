package model

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context/ctxhttp"
)

// Error type
type Error string

// Error implements Error interface
func (e Error) Error() string {
	return string(e)
}

// ErrNotLoggedIn returned when the user is not logged into CPPM yet
const ErrNotLoggedIn = Error("Not authorized. Make sure you log in and your account has the proper privileges")

// ErrPageTooSmall when paginated commands are givel a page size too small (<=0)
const ErrPageTooSmall = Error("Page size is too small")

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

// Reply object, generic version
type Reply map[string]json.RawMessage

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
	Items []Reply `json:"items"`
}

// WrappedReply returned by endpoints that return multiple objects
type wrappedReply struct {
	Embedded wrappedItems `json:"_embedded"`
	Links    halLinks     `json:"_links"`
}

// Exhaust a channel, dismiss all replies
func Exhaust(replies chan Reply) {
	for _ = range replies {
	}
}

// Generic function to perform a REST request
func rest(ctx context.Context, method Method, url, token string, query map[string]string, request, reply interface{}, skipVerify bool) error {
	var body io.Reader
	if request != nil {
		jsonBody, err := json.Marshal(request)
		if err != nil {
			return err
		}
		body = bytes.NewReader(jsonBody)
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
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/json")
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
		return err
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return ErrNotLoggedIn
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: REST Status %s", resp.Status)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	/*log.Print("HACK: Response body = ", string(respBody))*/
	return json.Unmarshal(respBody, reply)
}

// Flush all replies to the channel
func (w wrappedReply) flush(output chan Reply) {
	for _, item := range w.Embedded.Items {
		output <- item
	}
}

// Follow a paginated stream
func follow(ctx context.Context, method Method, url, token string, query map[string]string, request interface{}, skipVerify bool) (chan Reply, error) {
	result := make(chan Reply)
	errors := make(chan error)
	logger := log.New(os.Stderr, "", 0)
	go func() {
		defer close(result)
		reply := Reply{}
		if err := rest(ctx, method, url, token, query, request, &reply, skipVerify); err != nil {
			errors <- err
			return
		}
		// If it is not a list of embedded replies, but just a single
		// item, return it.
		rawEmbedded, ok := reply["_embedded"]
		if !ok {
			errors <- nil
			result <- reply
			return
		}
		// If embedded, flush the first batch of results
		wReply := wrappedReply{}
		if err := json.Unmarshal(rawEmbedded, &wReply.Embedded); err != nil {
			errors <- err
			return
		}
		rawLinks, ok := reply["_links"]
		if ok {
			if err := json.Unmarshal(rawLinks, &wReply.Links); err != nil {
				errors <- err
				return
			}
		}
		// Release the caller function, continue in background
		close(errors)
		wReply.flush(result)
		for {
			// Run new request from link provided by HAL
			url := wReply.Links.Next.Href
			if url == "" || wReply.Links.Next.Href == wReply.Links.Self.Href {
				return
			}
			if err := rest(ctx, GET, url, token, nil, nil, &wReply, skipVerify); err != nil {
				// Error channel is already closed. Just log the problem.
				logger.Print(err)
				return
			}
			wReply.flush(result)
		}
	}()
	err := <-errors
	if err != nil {
		go Exhaust(result)
		return nil, err
	}
	return result, nil
}
