package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

type shellCommand struct {
	name string
	args []string
}

func (c shellCommand) isZero() bool {
	return c.name == ""
}

func (c shellCommand) String() string {
	str := &strings.Builder{}
	str.WriteString(c.name)

	for _, arg := range c.args {
		str.WriteString(" ")
		switch {
		case arg == "", hasSpace(arg):
			fmt.Fprintf(str, `"%s"`, arg)
		default:
			str.WriteString(arg)
		}
	}

	return str.String()
}

func hasSpace(str string) bool {
	for _, r := range []rune(str) {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func (c shellCommand) run() (out results) {
	if c.isZero() {
		out = results{
			Err: fmt.Errorf("missing command"),
		}
		return out
	}

	start := time.Now()
	cmd := exec.Command(c.name, c.args...)

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	out = results{
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		Code:     getExitCode(err),
		Err:      err,
		Duration: jsonDuration(time.Now().Sub(start)),
	}
	return out
}

type noopReader struct{}

func (noopReader) Read([]byte) (int, error) {
	return 0, nil
}

func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	code := -1 // Default to a value that bash doesn't typically return

	execErr, ok := err.(*exec.ExitError)
	if ok {
		code = execErr.ExitCode()
	}
	return code
}
