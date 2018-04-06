package qiskit_api_go

import (
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"fmt"
	"sync"
	"time"
)

func init() {
	// Set up logger
	log.SetOutput(os.Stdout)
}

type clientOptions struct {
	// API User specific data
	clientAppl string

	// Job Execution stuff
	backend string
	shots int
	name string
	timeout time.Duration
	seed uint64
	maxCredits int
	mso bool	// HPC multi_shot_optimization
	omp int		// HPC omp_num_threads

	// IBM Q Info
	hub string
	group string
	project string
}

const (
	// DefaultClientAppl is the default client application name used by the custom HTTP header for the IBM QX API
	DefaultClientAppl = "qiskit-sdk-go"
	// DefaultMSO is the default HPC multi shot optimization value
	DefaultMSO = true
	// DefaultOMP is the default HPC omp number of threads value
	DefaultOMP = 16
)

// MaxSeed is the maximum seed value
const MaxSeed uint64 = 9999999999

// ClientOption configures how the client is set up
type ClientOption func(clientOptions)

// WithClientApplication specifies which client is using the QX Platform
func WithClientApplication(appl string) ClientOption {
	return func(options clientOptions) {
		options.clientAppl = DefaultClientAppl + ":" + appl
	}
}

// WithBackend
func WithBackend(backend string) ClientOption {
	return func(options clientOptions) {
		options.backend = backend
	}
}

// WithShots
func WithShots(shots int) ClientOption {
	return func(options clientOptions) {
		options.shots = shots
	}
}

// WithName
func WithName(name string) ClientOption {
	return func(options clientOptions) {
		options.name = name
	}
}

// JobTimeout
func JobTimeout(timeout time.Duration) ClientOption {
	return func(options clientOptions) {
		options.timeout = timeout
	}
}

// WithSeed configures the client to seed simulators before Jobs are ran with the given seed value
// Note: the seed value must be less than 11 digits long
func WithSeed(seed uint64) ClientOption {
	return func(options clientOptions) {
		options.seed = seed
	}
}

// WithMaxCredits
func WithMaxCredits(credits int) ClientOption {
	return func(options clientOptions) {
		options.maxCredits = credits
	}
}

// WithHPC configures the client to run jobs on the HPC simulator with the provided configuration values
// mso = multi_shot_optimization
// omp = omp_num_threads (must be between 1 and 16)
func WithHPC(mso bool, omp int) ClientOption {
	return func(options clientOptions) {
		options.mso = mso
		options.omp = omp
	}
}

// WithIbmQInfo configures the client to use the IBM Q features
func WithIbmQInfo(hub, group, project string) ClientOption {
	return func(options clientOptions) {
		options.hub = hub
		options.group = group
		options.project = project
	}
}

var maxQubitErrRegex = regexp.MustCompile(`.*register exceed the number of qubits, it can't be greater than (\d+).*`)

// Client represents a concurrent-safe IBM QX API client
// It implements the same methods as the python client so transferring shouldn't be difficult
type Client struct {
	mu sync.Mutex

	opts clientOptions
	conn *Conn
	backends map[string]*Backend
	jobs map[string]*Job
}

// NewClient returns a IBMQuantumExperience API Client
func NewClient(conn *Conn, options ...ClientOption) *Client {
	var opts clientOptions
	for _, option := range options {
		option(opts)
	}

	// Set defaults
	if opts.clientAppl == "" {
		opts.clientAppl = DefaultClientAppl
	}

	// Create client
	return &Client{
		opts: opts,
		conn: conn,
		backends: make(map[string]*Backend),
		jobs: make(map[string]*Job),
	}
}

// Version retrieves the current API version
func (c *Client) Version() float64 {
	resp, err := c.conn.get("version", "")
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	var i float64
	err = c.conn.decode(resp.Body, &i)
	if err != nil {
		panic(err)
	}

	return i
}

// Credit represents the users credits information
type Credit struct {
	MaxUserType float64	`json:"maxUserType,omitempty"`
	Promotional	float64	`json:"promotional,omitempty"`
	Remaining	float64	`json:"remaining,omitempty"`
}

type creditsResp struct {
	Err *httpErr	`json:"error,omitempty"`
	Cred Credit	`json:"credit,omitempty"`
}

// GetMyCredits returns the number of remaining credits associated with the given client
func (c *Client) GetMyCredits() Credit {
	resp, err := c.conn.get(fmt.Sprintf("users/%s", c.conn.dopts.userId), "")
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	var cResp creditsResp
	err = c.conn.decode(resp.Body, &cResp)
	if err != nil {
		log.Fatalln(err)
	}

	if cResp.Err != nil {
		log.Warn(cResp.Err)
	}

	return cResp.Cred
}

// Code represents a code
type Code struct {
	Name string				`json:"name,omitempty"`
	CreationDate string		`json:"creationDate,omitempty"`
	UserDeleted bool		`json:"userDeleted,omitempty"`
	UserId string			`json:"userId,omitempty"`
	Type string				`json:"type,omitempty"`
	Deleted bool			`json:"deleted,omitempty"`
	DisplayUrls map[string]string	`json:"displayUrls,omitempty"`
	IsPublic bool			`json:"isPublic,omitempty"`
	Id string				`json:"id,omitempty"`
	Qasm string				`json:"qasm,omitempty"`
	CodeType string			`json:"codeType,omitempty"`
	OrderDate float64		`json:"orderDate,omitempty"`
	Active bool				`json:"active,omitempty"`
	VersionId float64		`json:"versionId,omitempty"`
	IdCode string			`json:"idCode,omitempty"`
	Columns float64			`json:"numberColumns,omitempty"`
	Lines float64			`json:"numberLines,omitempty"`
	Gates float64			`json:"numberGates,omitempty"`
	HasMeasure bool			`json:"hasMeasure,omitempty"`
	Topology string			`json:"topology,omitempty"`
	HasBloch bool			`json:"hasBloch,omitempty"`
	GateDefs interface{}	`json:"gateDefinitions,omitempty"`
}

// GetCode retrieves a code by its id
func (c *Client) GetCode(codeId string) (code Code, err error) {
	resp, err := c.conn.get(fmt.Sprintf("Codes/%s", codeId), "")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = c.conn.decode(resp.Body, &code)
	return
}

// LatestCodes represents the latest codes associated with the user
type LatestCodes struct {
	Err 	*httpErr `json:"error,omitempty"`
	Total	float64	`json:"total,omitempty"`
	Count	float64	`json:"count,omitempty"`
	Codes 	[]Code `json:"codes,omitempty"`
}

// GetLastCodes returns the last codes of the user
func (c *Client) GetLastCodes() (LatestCodes, error) {
	resp, err := c.conn.get(fmt.Sprintf("users/%s/codes/latest", c.conn.dopts.userId), "&includeExecutions=true")
	if err != nil {
		log.Error(err)
		return LatestCodes{}, err
	}
	defer resp.Body.Close()

	var i LatestCodes
	err = c.conn.decode(resp.Body, &i)
	return i, err
}

// GetImageCode retrieves the image of a code, by its id
func (c *Client) GetImageCode(codeId string) (string, error) {
	resp, err := c.conn.get(fmt.Sprintf("Codes/%s/export/png/url", c.conn.dopts.accessToken), "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var i interface{}
	err = c.conn.decode(resp.Body, &i)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(i)
	return "", nil
}

// GetExecution retrieves an execution, by its ID
func (c *Client) GetExecution(executionId string) interface{} {
	resp, err := c.conn.get(fmt.Sprintf("Executions/%s", executionId), "")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var i interface{}
	err = c.conn.decode(resp.Body, &i)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(i)
	return i
}

// GetResultFromExecution retrieves the results of an execution, by its ID
func (c *Client) GetResultFromExecution(executionId string) {

}