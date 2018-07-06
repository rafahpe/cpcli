package model

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Authentication request
type authRequest map[string]string

// Authentication Reply
type authReply struct {
	Token   string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Expires int    `json:"expires_in"`
	Type    string `json:"token_type"`
	Scope   string `json:"scope"`
}

// Perform an authentication request
func (c *clearpass) auth(ctx context.Context, address string, req authRequest) (string, string, error) {
	baseURL := apiURL(address)
	fullURL := baseURL + "/oauth"
	rep := authReply{}
	if err := rest(ctx, c.client, POST, fullURL, "", nil, req, &rep); err != nil {
		return "", "", err
	}
	c.apiURL, c.webURL, c.token, c.refresh = baseURL, webURL(address), rep.Token, rep.Refresh
	return c.token, c.refresh, nil
}

// Login into clearpass with the provided credentials, return token.
func (c *clearpass) Login(ctx context.Context, address, clientID, secret, username, pass string) (string, string, error) {
	req := authRequest{
		"grant_type": "client_credentials",
		"client_id":  clientID,
	}
	if secret != "" {
		req["client_secret"] = secret
	}
	if username != "" && pass != "" {
		req["grant_type"] = "password"
		req["username"] = username
		req["password"] = pass
	}
	return c.auth(ctx, address, req)
}

// Validate the token is still useful
func (c *clearpass) Validate(ctx context.Context, address, clientID, secret, token, refresh string) (string, string, error) {
	// If there is arefresh token, try to refresh auth
	if refresh != "" {
		req := authRequest{
			"grant_type":    "refresh_token",
			"client_id":     clientID,
			"refresh_token": refresh,
			"client_secret": secret,
		}
		t, r, err := c.auth(ctx, address, req)
		if err == nil {
			return t, r, err
		}
	}
	// No refresh or it didn't succeed, just check the
	// current token is still valid.
	baseURL := apiURL(address)
	fullURL := baseURL + "/api-client/" + url.PathEscape(clientID)
	var rep RawReply
	if err := rest(ctx, c.client, GET, fullURL, token, nil, nil, &rep); err != nil {
		return "", "", err
	}
	c.apiURL, c.webURL, c.token, c.refresh = baseURL, webURL(address), token, ""
	return c.token, c.refresh, nil
}

// WebLogin into clearpass with the provided credentials, return cookies.
func (c *clearpass) WebLogin(ctx context.Context, address, username, pass string) ([]*http.Cookie, error) {
	baseURL := webURL(address)
	cookieURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	// Reset the cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Print("Error creating cookieJar, will not be able to use web login: ", err)
		jar = nil
	}
	c.client.Jar = jar
	// Get the first cookie
	fullURL := baseURL + "/tipsLogin.action"
	req, err := http.NewRequest(string(GET), fullURL, nil)
	if err != nil {
		return nil, err
	}
	detail, _ := rawRequest(ctx, c.client, req, nil, false)
	if detail.Err != nil {
		return nil, err
	}
	// Get the DWR cookie
	fullURL = baseURL + "/dwr/call/plaincall/__System.generateId.dwr"
	dwrBody := "callCount=1\nc0-scriptName=__System\nc0-methodName=generateId\nc0-id=0\nbatchId=0\ninstanceId=0\npage=%2Ftips%2FtipsLogin.action\nscriptSessionId=\n"
	req, err = http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	detail, _ = rawRequest(ctx, c.client, req, []byte(dwrBody), false)
	if detail.Err != nil {
		return nil, detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.New("Failed to get dwr cookie with status != 200")
		return nil, detail
	}
	// Parse the response and add the DWR cookie to the jar
	dwrCookie := ""
	parts := strings.SplitN(string(detail.Reply), "dwr.engine.remote.handleCallback(", 2)
	if len(parts) > 1 {
		parts = strings.SplitN(parts[1], "\"", 7)
		if len(parts) > 5 {
			cookies := c.client.Jar.Cookies(cookieURL)
			dwrCookie = parts[5]
			cookies = append(cookies, &http.Cookie{Name: dwrSessionCookie, Value: dwrCookie})
			jar.SetCookies(cookieURL, cookies)
		}
	}
	// Send the second pointless xhr request
	fullURL = baseURL + "/dwr/call/plaincall/beforeLogin.getPublisherUrl.dwr"
	dwrBody = "callCount=1\nnextReverseAjaxIndex=0\nc0-scriptName=beforeLogin\nc0-methodName=getPublisherUrl\nc0-id=0\nbatchId=1\ninstanceId=0\npage=%2Ftips%2FtipsLogin.action\nscriptSessionId=" + dwrCookie + "\n"
	req, err = http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	detail, _ = rawRequest(ctx, c.client, req, []byte(dwrBody), false)
	if detail.Err != nil {
		return nil, detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.New("Failed to confirm dwr cookie with status != 200")
		return nil, detail
	}
	// Post the login data
	fullURL = baseURL + "/tipsLoginSubmit.action"
	data := make(url.Values)
	data.Add("F_password", "0")
	data.Add("username", username)
	data.Add("password", pass)
	req, err = http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	detail, _ = rawRequest(ctx, c.client, req, []byte(data.Encode()), false)
	if detail.Err != nil {
		return nil, detail
	}
	// Finally, validate
	cookies, err := c.WebValidate(ctx, address)
	if err != nil {
		return nil, err
	}
	c.webURL, c.apiURL = baseURL, apiURL(address)
	return cookies, nil
}

// WebValidate checks the cookie is still useful
func (c *clearpass) WebValidate(ctx context.Context, address string) ([]*http.Cookie, error) {
	baseURL := webURL(address)
	fullURL := baseURL + "/tipsContent.action"
	req, err := http.NewRequest(string(GET), fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", baseURL+"/tipsLogin.action")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "deflate")
	req.Header.Add("Accept-Encoding", "br")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	detail, _ := rawRequest(ctx, c.client, req, nil, false)
	if detail.Err != nil {
		return nil, detail
	}
	cookies := c.cookies(baseURL)
	if detail.StatusCode != 200 || cookies == nil || len(cookies) < 2 {
		detail.Err = ErrNotLoggedIn
		return nil, detail
	}
	return cookies, nil
}

// WebLogout from clearpass.
// TODO: Check out. Does not seem to work.
func (c *clearpass) WebLogout(ctx context.Context, address string) error {
	baseURL := webURL(address)
	// Find dwr session cookie
	cookies := c.cookies(baseURL)
	if cookies == nil {
		return errors.New("Could not retrieve cookies")
	}
	dwrCookie := ""
	for _, cookie := range cookies {
		if cookie.Name == dwrSessionCookie {
			dwrCookie = cookie.Value
			break
		}
	}
	if cookies == nil {
		return errors.New("Could not retrieve DWR session cookie")
	}
	// Call XHR to close the session
	fullURL := baseURL + "/dwr/call/plaincall/login.destroySession.dwr"
	dwrBody := "callCount=1\nnextReverseAjaxIndex=0\nc0-scriptName=login\nc0-methodName=destroySession\nc0-id=0\nbatchId=1\ninstanceId=0\npage=%2Ftips%2FtipsContent.action\nscriptSessionId=" + dwrCookie + "\n"
	req, err := http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	detail, _ := rawRequest(ctx, c.client, req, []byte(dwrBody), false)
	if detail.Err != nil {
		return detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.New("Failed to close session with statuscode != 200")
		return detail
	}
	// Call to checkStatus to make the logout effective
	fullURL = baseURL + "/tipsLoginCheck.action"
	req, err = http.NewRequest(string(GET), fullURL, nil)
	if err != nil {
		return err
	}
	detail, _ = rawRequest(ctx, c.client, req, nil, false)
	if detail.Err != nil {
		return detail
	}
	if detail.StatusCode != 302 {
		detail.Err = errors.New("Did not get a redirect from tipsLoginCheck")
		return detail
	}
	return nil
}
