package model

import (
	"encoding/json"
	"strings"
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

// currentReply is the actual reply returned by rest calls
type currentReply struct {
	items     []RawReply
	lastError error
}

// reply is the implementation of the Reply iterator
type reply struct {
	curr   currentReply
	offs   int
	stream chan currentReply
}

// Next implements Reply interface
func (r *reply) Next() bool {
	// If offset is not at the end of r.curr.items, increment
	if r.curr.items != nil && r.offs < (len(r.curr.items)-1) {
		r.offs++
		return true
	}
	// Else, we exhausted the current block, go for the next one.
	// There won't be more replies if the flow is nil, or last reply was an error...
	if r.stream == nil || r.curr.lastError != nil {
		return false
	}
	// The current data is not an error, get a new one.
	r.curr = <-r.stream
	r.offs = 0
	// If the channel is closed, return false
	if r.curr.items == nil {
		r.stream = nil
		return false
	}
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
		curr: currentReply{
			items: []RawReply{r},
		},
		offs: -1, // So it is incremented to 0 when Next() is called
	}
}

// Pick particular attributes form a RawReply object
func (data RawReply) pick(attrib string) string {
	parts := strings.Split(attrib, ".")
	lenp := len(parts)
	// If the string is a dot-separated path, go deep
	for i := 0; i < lenp-1; i++ {
		newData, ok := data[parts[i]]
		if !ok {
			return ""
		}
		data = make(RawReply)
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
func (data RawReply) ToCSV(selectors []string) string {
	result := make([]string, 0, len(selectors))
	for _, attrib := range selectors {
		result = append(result, data.pick(attrib))
	}
	return strings.Join(result, ";")
}
