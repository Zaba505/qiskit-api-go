package qiskit_api_go

import (
	"testing"
	"os"
	"flag"
)

// These tests are to mimic the Python unit tests, as well as, test for concurrency safe-ness
// To run tests, run: go test -t YOUR_API_TOKEN

var (
	apiToken = flag.String("t", "", "Specifies the API token to use for the unit tests")
	testClient *Client
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *apiToken == "" {
		flag.Usage()
	}

	conn, err := Dial(WithApiToken(*apiToken))
	if err != nil {
		panic(err)
	}

	testClient = NewClient(conn)
	os.Exit(m.Run())
}

func TestClient_Version(t *testing.T) {
	v := testClient.Version()
	if v <= 4 {
		t.Fail()
	}
}

func TestClient_GetMyCredits(t *testing.T) {
	creds := testClient.GetMyCredits()
	if creds.Remaining <= 0 {
		t.Fail()
	}
}

func TestClient_GetLastCodes(t *testing.T) {
	codes, err := testClient.GetLastCodes()
	if err != nil {
		t.Error(err)
	}

	if codes.Codes == nil {
		t.Fail()
	}
}