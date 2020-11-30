package qiskit_api_go

import (
	"context"
	"time"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"bytes"
	"encoding/json"
	"strings"
)

var jobLogger = logrus.New()

const (
	// DefaultBackend is the default backend for Jobs and Experiments to be run on
	DefaultBackend = "simulator"
	// DefaultShots is the default number of shots a Experiment/Job can be ran for
	DefaultShots = 1
	// DefaultNameFmt is default Experiment name format to be used unless specified otherwise
	DefaultNameFmt = "Experiment #%d%d%d%d%d%d"
	// MaxShots is the maximum shots a experiment can be ran for
	MaxShots = 8192
	// MaxTimeout is the maximum timeout allowed for waiting on an experiment result
	MaxTimeout = 300 * time.Second
)

// Job represents one or more QASM 2.0 Experiments
type Job struct {
	// Some context shit
	mu sync.Mutex
	isExperiment bool

	// Id is the Jobs Id
	Id string	`json:"id,omitempty"`
	// Name is the name for this Job
	Name string	`json:"name,omitempty"`
	// Timeout is a timeout used by experiments
	Timeout time.Duration	`json:"timeout,omitempty"`
	// Shots is the number of shots ran
	Shots int	`json:"shots,omitempty"`
	// MaxCredits specifies the max credits to be used by this Job when executing
	MaxCredits int	`json:"maxCredits,omitempty"`
	// Qasm is all the qasm code to be executed by this Job
	Qasm []string	`json:"qasm,omitempty"`
}

// NewJob returns a Job which is a composition of experiments and specifications of how they should be executed
func NewJob(qasms []string, shots, maxCredits int) *Job {
	if shots > MaxShots {
		jobLogger.Warnf("shots were more than the maximum, %d, so they were set to be the maximum shots, %d", shots, MaxShots)
		shots = MaxShots
	}

	return &Job{Shots: shots, MaxCredits: maxCredits, Qasm: qasms}
}

// setId is a concurrent safe setter for the Jobs' Id
func (j *Job) setId(jobId string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Id = jobId
}

type jobExecReq struct {
	Qasm string		`json:"qasm,omitempty"`
	CodeType string	`json:"codeType,omitempty"`
	Name string		`json:"name,omitempty"`
	Qasms []string	`json:"qasms,omitempty"`
	Shots float64	`json:"shots,omitempty"`
	Bckend Backend	`json:"backend,omitempty"`
	MaxCredit float64	`json:"maxCredit,omitempty"`
	Seed int32	`json:"seed,omitempty"`
	Hpc	struct {
		MSO bool	`json:"multi_shot_optimization,omitempty"`
		OMP int		`json:"omp_num_threads,omitempty"`
	}	`json:"hpc,omitempty"`
}

type jobExecResp struct {
	Err *httpErr	`json:"error,omitempty"`

	Id string		`json:"id,omitempty"`
	DeviceId string	`json:"deviceId,omitempty"`
	Shots float64	`json:"shots,omitempty"`
	Deleted bool	`json:"deleted,omitempty"`
	ModDate float64	`json:"modificationDate,omitempty"`
	DeviceRunType string	`json:"deviceRunType,omitempty"`
	Time float64	`json:"time,omitempty"`
	EndDate string	`json:"endDate,omitempty"`
	InfoQueue interface{}	`json:"infoQueue,omitempty"`

	ParamsCustomize struct {
		Seed float64	`json:"seed,omitempty"`
	}	`json:"paramsCustomize,omitempty"`

	Status struct {
		Id string	`json:"id,omitempty"`
	}	`json:"status,omitempty"`

	Result expResp	`json:"result,omitempty"`
	Calib Calibration	`json:"calibration,omitempty"`
	Code Code	`json:"code,omitempty"`
}

// expResp represents the result returned by an experiment
type expResp struct {
	Date string	`json:"date,omitempty"`
	Data struct {
		P struct {
			Qubits []int	`json:"qubits,omitempty"`
			Labels []string	`json:"labels,omitempty"`
			Values []float64	`json:"values,omitempty"`
		}	`json:"p,omitempty"`
		AdditionalData struct {
			Seed float64	`json:"seed,omitempty"`
		}	`json:"additionalData,omitempty"`
		Qasm string		`json:"qasm,omitempty"`
		SerialNumDevice string	`json:"serialNumberDevice,omitempty"`
		Time float64	`json:"time,omitempty"`
		CregLabels string	`json:"creg_labels,omitempty"`
	}	`json:"data,omitempty"`
}

// ExpResult represents the result info to be returned by RunExperiment
type ExpResult struct {
	Status string	`json:"status,omitempty"`
	Id string	`json:"idExecution,omitempty"`
	CodeId string	`json:"idCode,omitempty"`
	InfoQueue interface{}	`json:"infoQueue,omitempty"`
	Result struct {
		ExtraInfo struct {
			Seed float64	`json:"seed,omitempty"`
		}	`json:"additionalData,omitempty"`
		Measure struct {
			Qubits []int	`json:"qubits,omitempty"`
			Labels []string	`json:"labels,omitempty"`
			Values []float64	`json:"values,omitempty"`
		}	`json:"measure,omitempty"`
		Bloch interface{}	`json:"bloch,omitempty"`
	}	`json:"result,omitempty"`
}

// RunExperiment runs the given shit as an experiment
func (c *Client) RunExperiment(ctx context.Context, qasm string, options ...ClientOption) error {
	// Set options
	for _, option := range options {
		option(c.opts)
	}

	// Set defaults
	if c.opts.backend == "" {
		c.opts.backend = DefaultBackend
	}
	if c.opts.name == "" {
		now := time.Now()
		c.opts.name = fmt.Sprintf(DefaultNameFmt, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	}
	if c.opts.shots == 0 {
		c.opts.shots = DefaultShots
	}

	// Check for a seed value
	if c.opts.seed > MaxSeed {
		return ApiErr{usrMsg: fmt.Sprintf("invalid seed (%d), seeds can have a maximum length of 10 digits", c.opts.seed)}
	}

	// Check backend
	backendType := c.checkBackend(c.opts.backend, "experiment")
	if backendType == "" {
		return BadBackendErr{backend: c.opts.backend}
	}

	// Tweak QASM
	qasm = strings.Replace(qasm, "IBMQASM 2.0;", "", -1)
	qasm = strings.Replace(qasm, "OPENQASM 2.0;", "", -1)

	// Construct parameters for the request
	var params string
	if c.opts.seed > 0 {
		params = fmt.Sprintf("&shots=%d&seed=%d&deviceRunType=%s", c.opts.shots, c.opts.seed, backendType)
	} else {
		params = fmt.Sprintf("&shots=%d&deviceRunType=%s", c.opts.shots, backendType)
	}

	// Create request body and send it
	var b bytes.Buffer
	req := &jobExecReq{
		Name: c.opts.name,
		Qasm: qasm,
		CodeType: "QASM2",
	}
	err := json.NewEncoder(&b).Encode(req)
	if err != nil {
		return err
	}

	resp, err := c.conn.post("codes/execute", params, &b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle response body
	var i jobExecResp
	err = c.conn.decode(resp.Body, &i)
	if err != nil {
		return err
	}

	if i.Err != nil {
		return i.Err
	}

	return nil
}

// RunJob runs the given job on the specified backend
func (c *Client) RunJob(ctx context.Context, j *Job, options ...ClientOption) error {
	// Set options
	for _, option := range options {
		option(c.opts)
	}

	// Set defaults
	if c.opts.backend == "" {
		WithBackend(DefaultBackend)(c.opts)
	}
	if c.opts.shots == 0 {
		WithShots(DefaultShots)(c.opts)
	}

	// Check for a seed value
	if c.opts.seed > MaxSeed {
		return ApiErr{usrMsg: fmt.Sprintf("invalid seed (%d), seeds can have a maximum length of 10 digits", c.opts.seed)}
	}

	// Check backend
	backendType := c.checkBackend(c.opts.backend, "job")
	if backendType == "" {
		return BadBackendErr{backend: c.opts.backend}
	}

	return nil
}

func (c *Client) GetJob(jobId string) {}
func (c *Client) GetJobs(jobIds ...string) {}
func (c *Client) CancelJob(jobId string) {}
