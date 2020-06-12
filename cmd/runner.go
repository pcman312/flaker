package cmd

import (
	"context"
	"fmt"

	"github.com/vrecan/life"
)

// runner executes shell commands and returns the results. This does no assertions, it is just a dumb parallel executor
type runner struct {
	*life.Life

	ctx context.Context

	shCmd   shellCommand
	results chan results
}

func newRunner(shCmd shellCommand, results chan results) (runner, error) {
	if results == nil {
		return runner{}, fmt.Errorf("results channel is nil")
	}

	r := runner{
		Life:    life.NewLife(),
		shCmd:   shCmd,
		results: results,
	}
	r.SetRun(r.run)
	return r, nil
}

func (r runner) run() {
	for {
		select {
		case <-r.Done:
			return
		default:
			r.results <- r.shCmd.run()
		}
	}
}
