package types

import (
	"fmt"
	"time"
)

type Results struct {
	StdOut   string       `json:"stdout,omitempty"`
	StdErr   string       `json:"stderr,omitempty"`
	Code     int          `json:"code"`
	Err      error        `json:"error,omitempty"`
	Duration JSONDuration `json:"duration"`
}

func (r Results) String() string {
	return fmt.Sprintf("Code: %d\n"+
		"Error: %s\n"+
		"\n"+
		"STDOUT:\n"+
		"%s\n"+
		"\n"+
		"STDERR:\n"+
		"%s\n",
		r.Code,
		r.Err,
		r.StdOut,
		r.StdErr)
}

type JSONDuration time.Duration

func (d JSONDuration) MarshalJSON() ([]byte, error) {
	b := []byte(fmt.Sprintf(`"%s"`, time.Duration(d)))
	return b, nil
}
