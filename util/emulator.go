package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// createRetryTest creates a bucket in the emulator and sets up a test using the
// Retry Test API for the given instructions. This is intended for emulator tests
// of retry behavior that are not covered by conformance tests.
func CreateRetryTest(host string, instructions map[string][]string) string {
	if len(instructions) == 0 {
		return ""
	}

	endpoint, err := url.Parse(host)
	if err != nil {
		fmt.Printf("Failed to parse host env: %v\n", err)
		os.Exit(0)
	}

	et := emulatorTest{name: "test", host: endpoint}
	et.create(instructions, "http")
	return et.id
}

type emulatorTest struct {
	name string
	id   string   // ID to pass as a header in the test execution
	host *url.URL // set the path when using; path is not guaranteed between calls
}

// Creates a retry test resource in the emulator
func (et *emulatorTest) create(instructions map[string][]string, transport string) {
	c := http.DefaultClient
	data := struct {
		Instructions map[string][]string `json:"instructions"`
		Transport    string              `json:"transport"`
	}{
		Instructions: instructions,
		Transport:    transport,
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		fmt.Printf("encoding request: %v\n", err)
	}

	et.host.Path = "retry_test"
	resp, err := c.Post(et.host.String(), "application/json", buf)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("creating retry test: err: %v, resp: %+v\n", err, resp)
		os.Exit(0)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	testRes := struct {
		TestID string `json:"id"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&testRes); err != nil {
		fmt.Printf("decoding test ID: %v\n", err)
	}

	et.id = testRes.TestID
	et.host.Path = ""
}

func (et *emulatorTest) delete() {
	et.host.Path = strings.Join([]string{"retry_test", et.id}, "/")
	c := http.DefaultClient
	req, err := http.NewRequest("DELETE", et.host.String(), nil)
	if err != nil {
		fmt.Errorf("creating request: %v", err)
		os.Exit(0)
	}
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Errorf("deleting test: err: %v, resp: %+v", err, resp)
	}
}
