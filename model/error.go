package model

import (
	"fmt"
	"net/http"
	"strings"
)

// Error type
type Error string

// Error implements Error interface
func (e Error) Error() string {
	return string(e)
}

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
