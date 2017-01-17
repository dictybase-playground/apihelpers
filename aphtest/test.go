// Package modwaretest provides common constants and functions for unit testing
package aphtest

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	// APIHost is the default http host for testing
	APIHost = "https://api.dictybase.org"
	// PathPrefix is the default prefix for appending to the API host
	PathPrefix = "1.0"
	// PubID is publication id for testing
	PubID = "99"
)

// IndentJSON uniformly indent the json byte
func IndentJSON(b []byte) []byte {
	var out bytes.Buffer
	_ = json.Indent(&out, b, "", " ")
	return bytes.TrimSpace(out.Bytes())
}

// APIServer returns a server URL
func APIServer() string {
	return fmt.Sprintf("%s/%s", APIHost, PathPrefix)
}

// MatchJSON compares actual and expected json
func MatchJSON(actual []byte, data interface{}) error {
	expected, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if bytes.Compare(IndentJSON(actual), IndentJSON(expected)) != 0 {
		return fmt.Errorf("actual %s and expected json %s are different", string(IndentJSON(actual)), string(IndentJSON(expected)))
	}
	return nil
}
