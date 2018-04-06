package qiskit_api_go

import "fmt"

// httpErr is an internal error container that is returned sometimes by the IBM QX API
type httpErr struct {
	Name 	 	string	`json:"name,omitempty"`
	Status 		int64	`json:"status,omitempty"`
	Message 	string	`json:"message,omitempty"`
	StatusCode 	int64	`json:"statusCode,omitempty"`
	Code 	   	string	`json:"code,omitempty"`
}
func (e *httpErr) Error() string { return fmt.Sprintf("name: %s status: %d message: %s statusCode: %d code: %s", e.Name, e.Status, e.Message, e.StatusCode, e.Code) }

type ApiErr struct {
	usrMsg, devMsg string
}
func (e ApiErr) Error() string { return fmt.Sprintf("usr_msg: %s dev_msg: %s", e.usrMsg, e.devMsg) }

type BadBackendErr struct {
	ApiErr
	backend string
}
func (e BadBackendErr) Error() string {
	e.usrMsg = fmt.Sprintf("could not find backend \"%s\" available", e.backend)
	e.devMsg = fmt.Sprintf("backend \"%s\" does not exist. please use client.AvailableBackends to get options", e.backend)
	return e.ApiErr.Error()
}

// CredentialsErr represents bad server credentials
type CredentialsErr struct {
	ApiErr
}

// RegisterSizeErr represents exceeding the maximum number of allowed qubits
type RegisterSizeErr struct {
	ApiErr
}