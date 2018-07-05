package model

import (
	"fmt"
	"net/http"
	"strings"
)

// Method is the HTTP method for the request
type Method string

// HTTP methods supported
const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	UPDATE Method = "UPDATE"
	DELETE Method = "DELETE"
	PATCH  Method = "PATCH"
)

// Error type
type Error string

// Error implements Error interface
func (e Error) Error() string {
	return string(e)
}

// RestError encodes info about REST errors
type RestError struct {
	Err         error
	Method      Method
	URL         string
	Query       string
	Header      http.Header
	Body        []byte
	StatusCode  int
	ReplyHeader http.Header
	Reply       []byte
}

// Error implements Error interface
func (e RestError) Error() string {
	return strings.Join([]string{
		e.Err.Error(),
		fmt.Sprint("Method: ", e.Method),
		fmt.Sprint("URL: ", e.URL),
		fmt.Sprint("Query: ", e.Query),
		fmt.Sprint("Header: ", e.Header),
		fmt.Sprint("Body: ", string(e.Body)),
		fmt.Sprint("StatusCode: ", e.StatusCode),
		fmt.Sprint("ReplyHeader: ", e.ReplyHeader),
		fmt.Sprint("Reply: ", string(e.Reply)),
	}, "\n  ")
}
