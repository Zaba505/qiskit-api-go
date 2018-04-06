package qiskit_api_go

import (
	"fmt"
	"log"
	"strings"
)

// OldBackends is a map of all the recognized old backend names
var OldBackendNames = map[string]string{
	"ibmqx5qv2": "real",
	"ibmqx2": "real",
	"qx5qv2": "real",
	"qx5q": "rea;",
	"real": "real",
	"ibmqx3": "ibmqx3",
	"simulator": "sim_trivial_2",
	"sim_trivial_2": "sim_trivial_2",
	"ibmqx_qasm_simulator": "sim_trivial_2",
}

// Backend represents a backend available to be used
type Backend struct {
	SerialNum	string	`json:"serialNumber,omitempty"`
	Id 			string	`json:"id,omitempty"`
	TopologyId	string	`json:"topologyId,omitempty"`
	CouplingMap	interface{}	`json:"couplingMap,omitempty"`	// Note: this is either 'all-to-all' or [][]int
	Name 		string	`json:"name,omitempty"`
	Status		string	`json:"status,omitempty"`
	Description	string	`json:"description,omitempty"`
	Simulator	bool	`json:"simulator,omitempty"`
	Nqubits		int64	`json:"nQubits,omitempty"`
	Version		float64	`json:"float64,omitempty"`
	OnlineDate	string	`json:"onlineDate,omitempty"`
	Url			string	`json:"url,omitempty"`
	ChipName	string	`json:"chipName,omitempty"`
	BasisGates	string	`json:"basisGates,omitempty"`
}

// Backends is an alias for a map of backend name to Backend data structure
type Backends map[string]*Backend

// Sims returns all the simulator backends out of this set of backends
func (bs Backends) Sims() (simBs []*Backend) {
	for _, b := range bs {
		if b.Simulator {
			simBs = append(simBs, b)
		}
	}
	return simBs
}

// AvailableBackends returns all the available backends that can be used
// If options is used it must be of length three and appear in this order: hub, group, project
func (c *Client) AvailableBackends(options ...ClientOption) Backends {
	for _, option := range options {
		option(c.opts)
	}

	var url string
	if c.opts.hub != "" && c.opts.group != "" && c.opts.project != "" {
		url = fmt.Sprintf("Network/%s/Groups/%s/Projects/%s/backends", c.opts.hub, c.opts.group, c.opts.project)
	} else {
		url = "Backends"
	}

	resp, err := c.conn.get(url, "")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var i []*Backend
	err = c.conn.decode(resp.Body, &i)
	if err != nil {
		log.Fatalln(err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, b := range i {
		if b.Status == "on" {
			c.backends[b.Name] = b
		}
	}

	return c.backends
}

func (c *Client) checkBackend(backendName, endpoint string) string {
	og_backend := backendName
	backendName = strings.ToLower(backendName)
	if endpoint == "experiment" {
		if b, exists := OldBackendNames[backendName]; exists {
			return b
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.backends[og_backend]; exists {
		return backendName
	}
	return ""
}

type calibErr struct {
	Date string	`json:"date,omitempty"`
	Value float64	`json:"value,omitempty"`
}

type paramsMeasure struct {
	Date string	`json:"date,omitempty"`
	Value float64	`json:"value,omitempty"`
	Unit string		`json:"unit,omitempty"`
}

// Status represents the status of a backend
type Status struct {
	Type string			`json:"backend,omitempty"`
	Available bool		`json:"state,omitempty"`
	Busy bool			`json:"busy,omitempty"`
	PendingJob int64	`json:"lengthQueue,omitempty"`
}

// TODO: Possibly wrap up Status, Calibration, and Parameters into one method
// BackendStatus retrieves the status of a chip
func (c *Client) BackendStatus(backend string) Status {
	backendType := c.checkBackend(backend, "status")
	if backendType == "" {
		log.Fatalf("unknown backend type: %s", backendType)
	}

	resp, err := c.conn.get(fmt.Sprintf("Backends/%s/queue/status", backendType), "withToken=false")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var r Status
	err = c.conn.decode(resp.Body, &r)
	if err != nil {
		log.Fatalln(err)
	}

	r.Type = backendType
	return r
}

func (c *Client) getBackendStatsUrl(backendType string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.opts.hub != "" {
		return fmt.Sprintf("Networks/%s/devices/%s", c.opts.hub, backendType)
	}
	return fmt.Sprintf("Backends/%s", backendType)
}

type Calibration struct {
	Type string			`json:"backend,omitempty"`
	LastUpdateDate string `json:"lastUpdateDate,omitempty"`
	MultiQubitGates []struct {
		Name    string   `json:"name,omitempty"`
		Type    string   `json:"type,omitempty"`
		Qubits  []int64  `json:"qubits,omitempty"`
		GateErr calibErr `json:"gateError,omitempty"`
	} `json:"multiQubitGates,omitempty"`
	Qubits []struct {
		Name       string   `json:"name,omitempty"`
		ReadOutErr calibErr `json:"readoutError,omitempty"`
		GateErr    calibErr `json:"gateError,omitempty"`
	}
}

// BackendCalibration retrieves the calibration of a chip
// The hub option is optional
func (c *Client) BackendCalibration(backend string, hub ClientOption) Calibration {
	if hub != nil {
		hub(c.opts)
	}

	backendType := c.checkBackend(backend, "calibration")
	if backendType == "" {
		log.Fatalf("unknown backend type: %s", backendType)
	}

	if backendType == "sim_trivial_2" {
		return Calibration{Type: backendType}
	}

	url := c.getBackendStatsUrl(backendType)
	resp, err := c.conn.get(url + "/calibration", "")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var h Calibration
	err = c.conn.decode(resp.Body, &h)
	if err != nil {
		log.Fatalln(err)
	}

	h.Type = backendType
	return h
}

// Params represents the calibration parameters for a backend
type Params struct {
	Type string			`json:"backend,omitempty"`

	FridgeParams struct {
		CooldownDate string		`json:"cooldownDate,omitempty"`
		Temp paramsMeasure		`json:"Temperature,omitempty"`
	}	`json:"fridgeParameters,omitempty"`

	Qubits []struct {
		Name string	`json:"name,omitempty"`
		GateTime paramsMeasure	`json:"gateTime,omitempty"`
		Freq paramsMeasure		`json:"frequency,omitempty"`
		T1 paramsMeasure		`json:"T1,omitempty"`
		T2 paramsMeasure		`json:"T2,omitempty"`
		Buffer paramsMeasure	`json:"buffer,omitempty"`
	}	`json:"qubits,omitempty"`
}

// BackendParameters retrieves the calibration parameters of a real chip
// The hub option is optional
func (c *Client) BackendParameters(backend string, hub ClientOption) Params {
	if hub != nil {
		hub(c.opts)
	}

	backendType := c.checkBackend(backend, "calibration")
	if backendType == "" {
		log.Fatalf("unknown backend type: %s", backendType)
	}

	if backendType == "sim_trivial_2" {
		return Params{Type: backendType}
	}

	url := c.getBackendStatsUrl(backendType)
	resp, err := c.conn.get(url + "/parameters", "")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var h Params
	err = c.conn.decode(resp.Body, &h)
	if err != nil {
		log.Fatalln(err)
	}

	return h
}