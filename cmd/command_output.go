package cmd

import (
	"fmt"
	"time"
)

type results struct {
	StdOut   string       `json:"stdout,omitempty"`
	StdErr   string       `json:"stderr,omitempty"`
	Code     int          `json:"code"`
	Err      error        `json:"error,omitempty"`
	Duration jsonDuration `json:"duration"`
}

func (co results) String() string {
	return fmt.Sprintf("Code: %d\n"+
		"Error: %s\n"+
		"\n"+
		"STDOUT:\n"+
		"%s\n"+
		"\n"+
		"STDERR:\n"+
		"%s\n",
		co.Code,
		co.Err,
		co.StdOut,
		co.StdErr)
}

type jsonDuration time.Duration

func (d jsonDuration) MarshalJSON() ([]byte, error) {
	b := []byte(fmt.Sprintf(`"%s"`, time.Duration(d)))
	return b, nil
}
