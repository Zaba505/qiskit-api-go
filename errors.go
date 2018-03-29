package qiskit_api_go

import "fmt"

type ApiErr struct {
	usrMsg, devMsg string
}
func (e ApiErr) Error() string { return fmt.Sprintf("usr_msg: %s\ndev_msg: %s", e.usrMsg, e.devMsg) }

func NewBadBackendErr(backend string) error {
	return ApiErr{
		fmt.Sprintf("Could not find backend \"%s\" available", backend),
		fmt.Sprintf("Backend \"%s\" does not exist. Please use client.AvailableBackends to see options", backend),
	}
}

// CredentialsErr represents bad server credentials
func NewCredentialsErr(usrMsg, devMsg string) error { return ApiErr{usrMsg, devMsg} }

// RegisterSizeErr represents exceeding the maximum number of allowed qubits
func NewRegisterSizeErr(usrMsg, devMsg string) error { return ApiErr{usrMsg, devMsg} }