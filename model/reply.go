package model

import (
	"context"
	"encoding/json"
	"strings"
)

// RawReply is the json body of each reply
type RawReply = json.RawMessage

// Reply is kind of an iterator on rest replies.
// you can iterate over this with a loop like:
// for r.Next(ctx) {
//   current := r.Get()
// }
// if r.Error() != nil {
//   iteration ended with error
// }
type Reply struct {
	current    []RawReply
	offset     int
	err        error
	token      string
	nextURL    string
	skipVerify bool
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

// Follow runs a REST request and returns an 'iterable' Reply
func follow(ctx context.Context, method Method, url, token string, query map[string]string, request interface{}, skipVerify bool) *Reply {
	r := &Reply{token: token, skipVerify: skipVerify, offset: -1}
	// Open the query and check for error.
	var reply RawReply
	if err := rest(ctx, method, url, token, query, request, &reply, skipVerify); err != nil {
		r.err = err
		return r
	}
	// If it is a plain reply (not wrapped), just return it straight away
	var wReply wrappedReply
	if err := json.Unmarshal(reply, &wReply); err != nil {
		r.current = []RawReply{reply}
		return r
	}
	// Otherwise, get the link for the next item
	r.current = wReply.Embedded.Items
	if nextURL := wReply.Links.Next.Href; nextURL != wReply.Links.Self.Href {
		r.nextURL = nextURL
	}
	return r
}

// Get returns the current reply
func (r *Reply) Get() RawReply {
	return r.current[r.offset]
}

// Next asks for the next reply in the stream
func (r *Reply) Next(ctx context.Context) bool {
	// If there is an error, stop iterating
	if r.err != nil {
		return false
	}
	// If there is still some data in the current list, move forward
	if r.current != nil && r.offset < (len(r.current)-1) {
		r.offset++
		return true
	}
	// Otherwise, keep asking for the next data
	for r.nextURL != "" {
		var wReply wrappedReply
		if err := rest(ctx, GET, r.nextURL, r.token, nil, nil, &wReply, r.skipVerify); err != nil {
			r.err = err
			return false
		}
		// Store the result and move offset to the first item
		r.current, r.offset, r.nextURL = wReply.Embedded.Items, 0, ""
		// And update the next URL
		if nextURL := wReply.Links.Next.Href; nextURL != wReply.Links.Self.Href {
			r.nextURL = nextURL
		}
		// If we got some fresh data, leave
		if len(r.current) > 0 {
			return true
		}
	}
	// If we reach here, we have no nextURL and no data
	return false
}

// Error returns the last error in the stream
func (r *Reply) Error() error {
	return r.err
}

// NewReply wraps a RawReply inside a Reply iterator
func NewReply(r RawReply, err error) *Reply {
	if err != nil {
		return &Reply{err: err}
	}
	return &Reply{
		current: []RawReply{r},
		offset:  -1,
	}
}

// Pick particular attributes from a RawReply object
func pick(data map[string]json.RawMessage, attrib string) string {
	parts := strings.Split(attrib, ".")
	lenp := len(parts)
	// If the string is a dot-separated path, go deep
	for i := 0; i < lenp-1; i++ {
		newData, ok := data[parts[i]]
		if !ok {
			return ""
		}
		if err := json.Unmarshal(newData, &data); err != nil {
			return ""
		}
	}
	result, ok := data[parts[lenp-1]]
	if !ok {
		return ""
	}
	repr, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(repr)
}

// ToCSV returns a line with the selected attribs of the object
func ToCSV(data RawReply, selectors []string) string {
	result := make([]string, 0, len(selectors))
	var mapdata map[string]json.RawMessage
	if err := json.Unmarshal(data, &mapdata); err != nil {
		return ""
	}
	for _, attrib := range selectors {
		result = append(result, pick(mapdata, attrib))
	}
	return strings.Join(result, ";")
}
