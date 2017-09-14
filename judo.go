// Package judo implements process spawning and management routines for simple
// and efficient forking of new processes from Go programs. It offers
// mechanisms to automatically kill long-running processes and manage the
// number of parallel running processes.
package judo

import (
	"time"

	"github.com/justsocialapps/justlib/work"
	"github.com/makkes/judo/process"
)

// A Spawner spawns and manages processes. It keeps track of running processes
// and kills them according to a given policy (e.g. maximum runtime).
type Spawner struct {
	worker    *work.Worker
	nextJobID uint64
}

type cmdParams struct {
	jobID    uint64
	quitChan chan int
	cmd      string
	argv     []string
}

// NewSpawner returns a new Spawner that can run maxProcs processes in
// parallel. It kills processes when they exceed a runtime of maxRuntime
// seconds.
func NewSpawner(maxProcs int, maxRuntime uint) Spawner {
	worker := work.NewWorker(maxProcs, func(payload work.Payload) interface{} {
		params := payload.Data.(cmdParams)
		quitChan := process.StartWithTimeout(params.cmd, params.argv, time.Duration(maxRuntime)*time.Second)
		return <-quitChan
	}, false)

	go func() {
		completions := worker.Completions()
		for {
			completion, ok := <-completions
			if !ok {
				return
			}
			inParams := completion.Input.(cmdParams)
			exitCode := completion.Output.(int)
			if inParams.quitChan != nil {
				inParams.quitChan <- exitCode
			}
		}
	}()

	return Spawner{
		worker: worker,
	}
}

// Quit stops the spawner from accepting new jobs to spawn and tears down all
// goroutines. All currently running processes are kept until either they quit
// or are killed.
func (s *Spawner) Quit() {
	s.worker.Quit()
}

// Spawn spawns a new process. The quitChan will receive the exit code of the
// process when it ends. This code is set to -1 when the process is killed
// after a timeout. If an error occurs, a non-nil error is returned.
func (s *Spawner) Spawn(cmd string, argv []string, quitChan chan int) error {
	jobID := s.nextJobID
	err := s.worker.Dispatch(work.Payload{
		Data: cmdParams{
			jobID:    jobID,
			quitChan: quitChan,
			cmd:      cmd,
			argv:     argv,
		},
	})
	if err != nil {
		return err
	}
	s.nextJobID++
	return nil
}
