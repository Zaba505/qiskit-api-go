package qiskit_api_go

import (
	"net/http"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"bytes"
	"io"
	"fmt"
	"strings"
)

func init() {
	// Set up logger
	log.SetOutput(os.Stdout)
}

var maxQubitErrRegex = regexp.MustCompile(`.*register exceed the number of qubits, it can't be greater than (\d+).*`)

type Client struct {
	opts *options
	c *http.Client
}

// NewClient returns a IBMQuantumExperience API Client
func NewClient(options ...ClientOption) (*Client, error) {
	opts := new(options)
	for _, option := range options {
		option(opts)
	}

	// Set defaults
	if opts.url == "" {
		opts.url = DefaultUrl
	}
	if opts.clientAppl == "" {
		opts.clientAppl = DefaultClientAppl
	}
	if opts.retries == 0 {
		opts.retries = DefaultRetries
	}
	if opts.timeout == 0 {
		opts.timeout = DefaultTimeout
	}

	// Create client
	c := &Client{
		opts: opts,
		c: &http.Client{
			Timeout: opts.timeout,
		},
	}

	// Get access token
	err := c.obtainToken()

	return c, err
}

func (c *Client) obtainToken() error {
	if c.opts.clientAppl == "" {
		c.opts.clientAppl = DefaultClientAppl
	}

	return nil
}

// TODO: Play around with the requests and get JSON responses so this can be done with JSON unmarshalling instead
// Do runs a http request
// This takes care of setting headers on requests also
// Note: This shouldn't be used by client but it is here to expose a little lower API if they want to
func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("x-qx-client-application", c.opts.clientAppl)
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	b := new(bytes.Buffer)
	retrys := c.opts.retries
	for retrys > 0 {
		b.Reset()

		// Execute the request
		resp, err = c.c.Do(req)
		if err != nil {
			return // TODO: Investigate this error
		}

		// Copy response body into buffer
		_, err = io.Copy(b, resp.Body)
		if err != nil {
			return // TODO: Investigate this error
		}

		// Check status code
		if resp.StatusCode != http.StatusOK {
			log.Warnf("Got a %d code response to %v", resp.StatusCode, resp.Request.URL)
			// TODO: Add something better than regex here
			if maxQubitErrRegex.MatchReader(b) {
				r := maxQubitErrRegex.FindAllStringSubmatch(string(b.Bytes()), -1)
				return nil, NewRegisterSizeErr(fmt.Sprintf("device register must be <= %s", r[0][1]), "")
			}
		}

		// Check Content-Type
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html;") {
			// TODO: Return
			return
		} else {

		}

		retrys--
	}

	return
}