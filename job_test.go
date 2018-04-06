package qiskit_api_go

import (
	"testing"
	"context"
)

const testExpStr = `IBMQASM 2.0;

include "qelib1.inc";
qreg q[5];
creg c[5];
u2(-4*pi/3,2*pi) q[0];
u2(-3*pi/2,2*pi) q[0];
u3(-pi,0,-pi) q[0];
u3(-pi,0,-pi/2) q[0];
u2(pi,-pi/2) q[0];
u3(-pi,0,-pi/2) q[0];
measure q -> c;`

func TestClient_RunExperiment(t *testing.T) {
	err := testClient.RunExperiment(context.Background(), testExpStr)
	if err != nil {
		t.Error(err)
	}
}

func TestClient_RunJob(t *testing.T) {}
func TestClient_RunJob_With_Seed(t *testing.T) {}
func TestClient_RunJob_Fail_Backend(t *testing.T) {}

func TestClient_GetJobs(t *testing.T) {}