package qiskit_api_go

import (
	"testing"
)

func TestClient_AvailableBackends(t *testing.T) {
	backends := testClient.AvailableBackends()
	if len(backends) < 2 {
		t.Fail()
	}

	t.Run("backend_sims", func(t2 *testing.T) {
		if len(backends.Sims()) < 1 {
			t2.Fail()
		}
	})
}

func TestClient_BackendStatus(t *testing.T) {
	status := testClient.BackendStatus("ibmqx4")
	if status.Type != "ibmqx4" {
		t.Fail()
	}
}

func TestClient_BackendCalibration(t *testing.T) {
	calibration := testClient.BackendCalibration("ibmqx4", nil)
	if calibration.MultiQubitGates == nil {
		t.Fail()
	}
}

func TestClient_BackendParameters(t *testing.T) {
	params := testClient.BackendParameters("ibmqx4", nil)
	if params.Qubits == nil {
		t.Fail()
	}
}