package results

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pcman312/flaker/cmd/types"
	"github.com/vrecan/life"
)

type ResultsListenerOpt func(*resultsListener) error

func ResultsChan(results chan types.Results) ResultsListenerOpt {
	return func(rl *resultsListener) error {
		rl.results = results
		return nil
	}
}

type stats interface {
	Record(output types.Results)
}

func Stats(state stats) ResultsListenerOpt {
	return func(rl *resultsListener) error {
		rl.state = state
		return nil
	}
}

func Writer(writer io.Writer) ResultsListenerOpt {
	return func(rl *resultsListener) error {
		rl.resultsWriter = writer
		return nil
	}
}

func StopOnFailure(stopFunc func()) ResultsListenerOpt {
	return func(rl *resultsListener) error {
		rl.stopOnce = &sync.Once{}
		rl.stopFunc = stopFunc
		return nil
	}
}

type resultsListener struct {
	*life.Life

	results       chan types.Results
	state         stats
	resultsWriter io.Writer

	stopOnce *sync.Once
	stopFunc func()
}

func NewListener(opts ...ResultsListenerOpt) (*resultsListener, error) {
	rl := &resultsListener{
		Life: life.NewLife(),
	}
	rl.SetRun(rl.run)

	merr := &multierror.Error{}

	for _, opt := range opts {
		merr = multierror.Append(merr, opt(rl))
	}

	if rl.results == nil {
		merr = multierror.Append(merr, fmt.Errorf("missing results channel"))
	}
	if rl.state == nil {
		merr = multierror.Append(merr, fmt.Errorf("missing state"))
	}

	return rl, merr.ErrorOrNil()
}

func (rl *resultsListener) run() {
	for {
		select {
		case <-rl.Done:
			return
		case result, more := <-rl.results:
			if !more {
				return
			}
			rl.state.Record(result)
			rl.writeResultsToFile(result)
			// If stopFunc is specified and a failure is detected, run the stop func
			if rl.stopFunc != nil && (result.Code != 0 || result.Err != nil) {
				rl.stopOnce.Do(rl.stopFunc)
			}
		}
	}
}

func (rl *resultsListener) writeResultsToFile(out types.Results) {
	if rl.resultsWriter == nil {
		return
	}

	b, err := json.Marshal(out)
	if err != nil {
		// TODO: Figure out where to record errors other than on the console
		fmt.Fprintf(os.Stderr, "ERROR marshalling results to JSON: %s\n", err)
	}

	b = append(b, '\n')
	_, err = rl.resultsWriter.Write(b)
	if err != nil {
		// TODO: Figure out where to record errors other than on the console
		fmt.Fprintf(os.Stderr, "ERROR writing results: %s\n", err)
	}
}
