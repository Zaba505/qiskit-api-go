package qiskit_api_go

import (
	"time"
	"net/http"
	"bytes"
	"encoding/json"
	"io"
	"fmt"
)

const (
	// DefaultUrl is the default IBM QX API Endpoint URL
	DefaultUrl = "https://quantumexperience.ng.bluemix.net/api"
	// DefaultRetries is the default number of retries every request gets
	DefaultRetries = 5
	// DefaultTimeout is the default timeout for each request
	DefaultTimeout = 30 * time.Second
)

type dialOptions struct {
	// Login Info
	apiToken string
	email string
	password string
	accessToken string
	userId string

	// API Endpoint Info
	url string
	proxyUrls map[string]string
	ntmlUsername string
	ntmlPassword string

	// API Request Info
	retries int
	timeout time.Duration
}

// DialOption configures how to connection works
type DialOption func(*dialOptions)

// WithApiToken configures the connection to obtain your access token by using your API token
func WithApiToken(token string) DialOption {
	return func(options *dialOptions) {
		options.apiToken = token
	}
}

// WithAccessInfo configures the connection already with an API Access Token and a User ID
func WithAccessInfo(token, userId string) DialOption {
	return func(options *dialOptions) {
		options.accessToken = token
		options.userId = userId
	}
}

// WithLoginInfo configures the connection to obtain your access token by using your login info
func WithLoginInfo(email, password string) DialOption {
	return func(options *dialOptions) {
		options.email = email
		options.password = password
	}
}

// WithApiUrl configures the connection to use the provided url for the API endpoints
func WithApiUrl(url string) DialOption {
	return func(options *dialOptions) {
		options.url = url
	}
}

// WithProxies configures the conn proxy information
// urls should be a map of:
//		http: URL
//		https: URL
// ntmlInfo should be length 2 where first value is username and second value is the password for NTML Auth
func WithProxies(urls map[string]string, ntmlInfo ...string) DialOption {
	return func(options *dialOptions) {
		options.proxyUrls = urls

		if len(ntmlInfo) == 2 {
			options.ntmlUsername = ntmlInfo[0]
			options.ntmlPassword = ntmlInfo[1]
		}
	}
}

// WithRetries configures the number of retries performed for any request
func WithRetries(retries int) DialOption {
	return func(options *dialOptions) {
		options.retries = retries
	}
}

// WithTimeout configures the timeout for each request
func WithTimeout(timeout time.Duration) DialOption {
	return func(options *dialOptions) {
		options.timeout = timeout
	}
}

// Conn is a representation of a connection to the IBM QX API
type Conn struct {
	dopts dialOptions
	c *http.Client
}

// Dial takes a list of DialOptions and returns a connection to the IBM QX API
func Dial(options ...DialOption) (*Conn, error) {
	c := &Conn{
		c: &http.Client{},
	}

	for _, option := range options {
		option(&c.dopts)
	}

	// Check API Login info; otherwise, error
	if c.dopts.apiToken == "" && c.dopts.email == "" && c.dopts.accessToken == "" {
		return nil, CredentialsErr{ApiErr{usrMsg: "missing credentials to obtain access token. please provide either, api token or email/password"}}
	}

	// Set defaults
	if c.dopts.url == "" {
		c.dopts.url = DefaultUrl
	}

	if c.dopts.retries == 0 {
		c.dopts.retries = DefaultRetries
	}

	if c.dopts.timeout == 0 {
		c.dopts.timeout = DefaultTimeout
	}
	c.c.Timeout = c.dopts.timeout

	// Lastly, obtain access token
	var err error
	if c.dopts.accessToken == "" {
		err = c.obtainToken()
	}
	return c, err
}

// loginReq is an internal type for making obtainToken requests
type loginReq struct {
	Token 		string	`json:"apiToken,omitempty"`
	Email 		string	`json:"email,omitempty"`
	Password 	string	`json:"password,omitempty"`
}

type loginResp struct {
	httpErr
	Created string `json:"created"`
	UserId string `json:"userId"`
	Id string	`json:"id"`
	Ttl	float64	`json:"ttl"`
}

func (c *Conn) obtainToken() error {
	// Construct request
	loginReq := loginReq{}
	switch {
	case c.dopts.apiToken != "":
		loginReq.Token = c.dopts.apiToken
	case c.dopts.email != "" && c.dopts.password != "":
		loginReq.Email = c.dopts.email
		loginReq.Password = c.dopts.password
	default:
		return CredentialsErr{ApiErr{usrMsg: "invalid credentials, please provide either API token or user email and password"}}
	}

	// Encode JSON request body
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(loginReq)
	if err != nil {
		return err
	}

	// Construct request URL
	url := c.dopts.url + "/users/login"
	if loginReq.Token != "" {
		url += "WithToken"
	}

	// Create request and execute it
	req, _ := http.NewRequest(http.MethodPost, url, &b)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle response
	var r loginResp
	err = c.decode(resp.Body, &r)
	if err != nil {
		return err
	}

	// Set fields
	c.dopts.userId = r.UserId
	c.dopts.accessToken = r.Id

	return nil
}

// newRequest is simply just a helper for generating requests
func (c *Conn) newRequest(method, path, params string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s?access_token=%s%s", c.dopts.url, path, c.dopts.accessToken, params), body)
	if err != nil {
		panic(err) // TODO: Implement better logging
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// decode is simply a helper for decoding json
func (c *Conn) decode(r io.Reader, i interface{}) (err error) {
	err = json.NewDecoder(r).Decode(i)
	return
}

// TODO: Implement better error handling shit
// Do runs a http request
// This takes care of setting headers on requests also
// Note: This shouldn't be used by client but it is here to expose a little lower API if they want to
func (c *Conn) do(req *http.Request) (resp *http.Response, err error) {
	retrys := c.dopts.retries
	for retrys > 0 {
		// Execute the request
		resp, err = c.c.Do(req)
		if err != nil {
			return // TODO: Investigate this error
		}

		// Check for 401 and get new token
		if resp.StatusCode == http.StatusUnauthorized {
			if err = c.obtainToken(); err != nil {
				return
			}

			resp, err = c.c.Do(req)
		}

		// Check status code
		if resp.StatusCode != http.StatusOK {
//			log.Warnf("Got a %d code response to %v", resp.StatusCode, resp.Request.URL)
			// TODO: Add something better than regex here
		} else {
			return
		}

		retrys--
	}

	err = ApiErr{usrMsg: "Failed to get proper response from backend"}
	return
}

// Post is a convenience wrapper around a POST request
func (c *Conn) post(path, params string, body io.Reader) (*http.Response, error) {
	req := c.newRequest(http.MethodPost, path, params, body)
	return c.do(req)
}

// Put is a convenience wrapper around a PUT request
func (c *Conn) put(path, params string, body io.Reader) (*http.Response, error) {
	req := c.newRequest(http.MethodPut, path, params, body)
	return c.do(req)
}

// Get is a convenience wrapper around a GET request
func (c *Conn) get(path, params string) (*http.Response, error) {
	req := c.newRequest(http.MethodGet, path, params, nil)
	return c.do(req)
}