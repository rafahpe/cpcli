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
	"strings"

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

// RestError encodes info about REST errors
type RestError struct {
	Err        error
	Method     string
	URL        string
	Query      string
	Unsafe     bool
	Header     http.Header
	Body       string
	StatusCode int
	Reply      string
}

// Error implements Error interface
func (e RestError) Error() string {
	return strings.Join([]string{
		e.Err.Error(),
		fmt.Sprint("Method: ", e.Method),
		fmt.Sprint("URL: ", e.URL),
		fmt.Sprint("Query: ", e.Query),
		fmt.Sprint("Unsafe: ", e.Unsafe),
		fmt.Sprint("Header: ", e.Header),
		fmt.Sprint("Body: ", e.Body),
		fmt.Sprint("StatusCode: ", e.StatusCode),
		fmt.Sprint("Reply: ", e.Reply),
	}, "\n  ")
}

// HTTP methods supported
const (
	GET    Method = "GET"
	POST          = "POST"
	PUT           = "PUT"
	UPDATE        = "UPDATE"
	DELETE        = "DELETE"
	PATCH         = "PATCH"
)

// RawReply is the json body of each reply
type RawReply map[string]json.RawMessage

// Reply object, "sort of" iterable object that allows iteration.
// you can iterate over this with a loop like:
// for r.Next() {
//   current := r.Get()
// }
// if r.Error() != nil {
//   iteration ended with error
// }
type Reply interface {
	Next() bool    // Next returns true if there is a reply to Get
	Get() RawReply // Get the current reply
	Error() error  // Error returns the non-nil error if the iteration broke
}

// replyData is the actual reply returned by rest calls
type replyData struct {
	items     []RawReply
	lastError error
}

// reply is the implementation of the Reply iterator
type reply struct {
	flow chan replyData
	curr replyData
	offs int
}

// Next implements Reply interface
func (r *reply) Next() bool {
	// If curr is empty or offs has reached the last item, get a new reply
	if r.curr.items == nil || r.offs >= (len(r.curr.items)-1) {
		// There won't be more replies if the flow is nil, or last reply was an error...
		if r.flow == nil || r.curr.lastError != nil {
			return false
		}
		// The current data is not an error, get a new one.
		r.curr = <-r.flow
		r.offs = 0
		// If the channel is closed, return false
		if r.curr.items == nil {
			r.flow = nil
			return false
		}
		return true
	}
	r.offs++
	return true
}

// Error implements Reply interface
func (r *reply) Error() error {
	return r.curr.lastError
}

// Get implements Reply interface
func (r *reply) Get() RawReply {
	return r.curr.items[r.offs]
}

// NewReply wraps a RawReply inside a Reply iterator
func NewReply(r RawReply) Reply {
	return &reply{
		curr: replyData{
			items: []RawReply{r},
		},
		offs: -1, // So it is incremented to 0 when Next() is called
	}
}

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
	result := make(chan replyData)
	waitme := make(chan error)
	go func() {
		defer close(result)
		reply := RawReply{}
		if err := rest(ctx, method, url, token, query, request, &reply, skipVerify); err != nil {
			waitme <- err
			return
		}
		// If it is not a list of embedded replies, but just a single
		// item, return it.
		rawEmbedded, ok := reply["_embedded"]
		if !ok {
			close(waitme)
			result <- replyData{items: []RawReply{reply}}
			return
		}
		// If embedded, flush the first batch of results
		wReply := wrappedReply{}
		if err := json.Unmarshal(rawEmbedded, &wReply.Embedded); err != nil {
			waitme <- err
			return
		}
		rawLinks, ok := reply["_links"]
		if ok {
			if err := json.Unmarshal(rawLinks, &wReply.Links); err != nil {
				waitme <- err
				return
			}
		}
		// Release the caller
		close(waitme)
		result <- replyData{items: wReply.Embedded.Items}
		for {
			// Run new request from link provided by HAL
			url := wReply.Links.Next.Href
			if url == "" || wReply.Links.Next.Href == wReply.Links.Self.Href {
				return
			}
			if err := rest(ctx, GET, url, token, nil, nil, &wReply, skipVerify); err != nil {
				result <- replyData{lastError: err}
				return
			}
			result <- replyData{items: wReply.Embedded.Items}
		}
	}()
	if err := <-waitme; err != nil {
		return nil, err
	}
	return &reply{flow: result}, nil
}
