package model

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const sessionCookie = "JSESSIONID"
const dwrSessionCookie = "DWRSESSIONID"

// ErrPageTooSmall when paginated commands are givel a page size too small (<=0)
const ErrPageTooSmall = Error("Page size is too small")

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM. Returns access and refresh tokens, or error.
	// - 'address' is the CPPM server address (:port, if different from 443).
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret. Empty if client is public (trusted).
	// - 'user', 'pass' are the username and password for "password"
	//   authentication. If any of them is empty,  'client_credentials'
	//   authentication is used instead.
	Login(ctx context.Context, address, clientID, secret, user, pass string) (string, string, error)
	// Validate / Refresh credentials.
	// - 'address' is the CPPM server address (:port, if different from 443).
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret.
	// - 'token', 'refresh': the authentication and refresh tokens.
	//   If a refresh token is provided, attempt to refresh the
	//   authentication token. Otherwise, just check it is valid.
	Validate(ctx context.Context, address, clientID, secret, token, refresh string) (string, string, error)
	// Token obtained after authentication / validation
	Token() string
	// WebLogin into clearpass with the provided credentials, return token.
	WebLogin(ctx context.Context, address, username, pass string) ([]*http.Cookie, error)
	// WebLogout from clearpass.
	WebLogout(ctx context.Context, address string) error
	// WebValidate checks the cookie is still useful
	WebValidate(ctx context.Context, address string) ([]*http.Cookie, error)
	// Cookies obtained after web authentication
	Cookies() []*http.Cookie
	// Request made to the CPPM.
	Request(method Method, path string, params Params, request interface{}) *Reply
	// Export some resource from ClearPass, return the exported stream.
	Export(ctx context.Context, resource, pass string) (string, io.ReadCloser, error)
	// Import some resource to ClearPass.
	Import(ctx context.Context, fileName, resource, pass string) error
}

// Clearpass model
type clearpass struct {
	unsafe  bool
	apiURL  string
	webURL  string
	token   string
	refresh string
	client  *http.Client
}

// apiURL returns the URL of the API
func apiURL(address string) string {
	return fmt.Sprintf("https://%s/api", url.PathEscape(address))
}

// webURL returns the URL of the web interface
func webURL(address string) string {
	return fmt.Sprintf("https://%s/tips", url.PathEscape(address))
}

// New creates a Clearpass object with cached IP and token
func New(address, token, refresh string, cookies []*http.Cookie, skipVerify bool) Clearpass {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Print("Error creating cookieJar, will not be able to use web login: ", err)
		jar = nil
	}
	queryURL := webURL(address)
	if cookies != nil && len(cookies) > 0 {
		if q, err := url.Parse(queryURL); err != nil {
			log.Print("Error parsing url, will not be able to use web login: ", err)
		} else {
			jar.SetCookies(q, cookies)
		}
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}
	return &clearpass{
		apiURL:  apiURL(address),
		webURL:  queryURL,
		token:   token,
		refresh: refresh,
		client:  client,
	}
}

// Token implements Clearpass interface
func (c *clearpass) Token() string {
	return c.token
}

// Cookie implements Clearpass interface
func (c *clearpass) Cookies() []*http.Cookie {
	return c.cookies(c.webURL)
}

func (c *clearpass) cookies(cpURL string) []*http.Cookie {
	queryURL, err := url.Parse(cpURL)
	if err != nil {
		return nil
	}
	cookies := c.client.Jar.Cookies(queryURL)
	for _, cookie := range cookies {
		// Check that at least the session cookie exists
		if strings.Compare(cookie.Name, sessionCookie) == 0 {
			return cookies
		}
	}
	return nil
}

// Follow a stream of results from an endpoint.
func (c *clearpass) Request(method Method, path string, params Params, request interface{}) *Reply {
	if c.apiURL == "" || c.token == "" {
		return NewReply(nil, ErrNotLoggedIn)
	}
	// Clone params, if any
	var defaults Params
	if params != nil && len(params) > 0 {
		defaults := make(Params)
		for k, v := range params {
			defaults[k] = v
		}
		if _, ok := defaults["limit"]; ok {
			defaults["calculate_count"] = "false"
		}
		if filter, ok := defaults["filter"]; ok {
			norm, err := normalize(filter, path)
			if err != nil {
				return NewReply(nil, err)
			}
			defaults["filter"] = norm
		}
	}
	return Request(c.client, method, c.apiURL+"/"+path, c.token, defaults, request)
}
