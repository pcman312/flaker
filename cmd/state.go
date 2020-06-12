package cmd

import (
	"sync"
)

type runState struct {
	mu          *sync.RWMutex
	successful  int64
	failed      int64
	computeTime int64
}

type status struct {
	runs        int64
	successful  int64
	failed      int64
	computeTime int64
}

func newRunState() *runState {
	s := &runState{
		mu: &sync.RWMutex{},
	}
	return s
}

func (s *runState) status() status {
	s.mu.RLock()
	successful := s.successful
	failed := s.failed
	computeTime := s.computeTime
	s.mu.RUnlock()

	snap := status{
		runs:        successful + failed,
		successful:  successful,
		failed:      failed,
		computeTime: computeTime,
	}
	return snap
}

func (s *runState) record(output results) {
	s.mu.Lock()
	if output.Code != 1 {
		s.successful++
	} else {
		s.failed++
	}
	dur := int64(output.Duration)
	s.computeTime += dur
	s.mu.Unlock()
}
