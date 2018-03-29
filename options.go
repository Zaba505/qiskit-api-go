package qiskit_api_go

import "time"

type options struct {
	token string
	url string
	accessToken string
	userId string
	clientAppl string
	proxyUrls map[string]string
	ntmlUsername string
	ntmlPassword string
	email string
	password string

	retries int
	timeout time.Duration
}

const (
	DefaultUrl = "https://quantumexperience.ng.bluemix.net/api"
	DefaultClientAppl = "qiskit-sdk-go"
	DefaultRetries = 5
	DefaultTimeout = 30 * time.Second
)

// ClientOption configures how the client is set up
type ClientOption func(*options)

// WithApiToken provides a ClientOption that sets the users API token
func WithApiToken(token string) ClientOption {
	return func(options *options) {
		options.token = token
	}
}

// WithApiUrl configures the client to use the provided url for the API endpoints
func WithApiUrl(url string) ClientOption {
	return func(options *options) {
		options.url = url
	}
}

// WithAccessToken sets the access token
func WithAccessToken(token string) ClientOption {
	return func(options *options) {
		options.accessToken = token
	}
}

// WithUserId sets the user id
func WithUserId(id string) ClientOption {
	return func(options *options) {
		options.userId = id
	}
}

// WithClientApplication specifies which client is using the QX Platform
func WithClientApplication(appl string) ClientOption {
	return func(options *options) {
		options.clientAppl = appl
	}
}

// WithProxies configures the client proxy information
// urls should be a map of:
//		http: URL
//		https: URL
// ntmlInfo should be length 2 where first value is username and second value is the password for NTML Auth
func WithProxies(urls map[string]string, ntmlInfo ...string) ClientOption {
	return func(options *options) {
		options.proxyUrls = urls

		if len(ntmlInfo) == 2 {
			options.ntmlUsername = ntmlInfo[0]
			options.ntmlPassword = ntmlInfo[1]
		}
	}
}

// WithLoginInfo configures the client to obtain your access token by using your login info
func WithLoginInfo(email, password string) ClientOption {
	return func(options *options) {
		options.email = email
		options.password = password
	}
}

// WithRetries configures the number of retries performed for any request
func WithRetries(retries int) ClientOption {
	return func(options *options) {
		options.retries = retries
	}
}

// WithTimeout configures the timeout for each request
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *options) {
		options.timeout = timeout
	}
}