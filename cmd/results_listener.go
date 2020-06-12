package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/vrecan/life"
)

type resultsListener struct {
	*life.Life

	results       chan results
	state         *runState
	resultsWriter io.Writer
}

func newResultsListener(results chan results, state *runState, resultsWriter io.Writer) (resultsListener, error) {
	if results == nil {
		return resultsListener{}, fmt.Errorf("results channel is nil")
	}
	if state == nil {
		return resultsListener{}, fmt.Errorf("state is nil")
	}

	rl := resultsListener{
		Life:          life.NewLife(),
		results:       results,
		state:         state,
		resultsWriter: resultsWriter,
	}
	rl.SetRun(rl.run)
	return rl, nil
}

func (rl resultsListener) run() {
	for {
		select {
		case <-rl.Done:
			return
		case out, more := <-rl.results:
			if !more {
				return
			}
			rl.state.record(out)
			rl.writeResultsToFile(out)
		}
	}
}

func (rl resultsListener) writeResultsToFile(out results) {
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
