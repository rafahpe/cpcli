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
	detail := make([]string, 0, 16)
	if e.Err != nil {
		detail = append(detail, e.Err.Error())
	}
	detail = append(detail, fmt.Sprint("Method: ", e.Method))
	detail = append(detail, fmt.Sprint("URL: ", e.URL))
	if e.Query != "" {
		detail = append(detail, fmt.Sprint("Query: ", e.Query))
	}
	if e.Header != nil {
		detail = append(detail, fmt.Sprint("Header: ", e.Header))
	}
	if e.Body != nil {
		detail = append(detail, fmt.Sprint("Body: ", string(e.Body)))
	}
	if e.StatusCode != 0 {
		detail = append(detail, fmt.Sprint("StatusCode: ", e.StatusCode))
	}
	if e.ReplyHeader != nil {
		detail = append(detail, fmt.Sprint("ReplyHeader: ", e.ReplyHeader))
	}
	if e.Reply != nil {
		detail = append(detail, fmt.Sprint("Reply: ", string(e.Reply)))
	}
	return strings.Join(detail, "\n  ")
}
