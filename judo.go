package judo

import (
	"time"

	"github.com/justsocialapps/justlib/work"
	"github.com/makkes/judo/process"
	log "gopkg.in/justsocialapps/justlib.v1/logging"
)

// A Spawner spawns and manages processes. It keeps track of running processes
// and kills them according to a given policy (e.g. maximum runtime).
type Spawner struct {
	worker *work.Worker
	jobCnt uint64
}

type cmdParams struct {
	jobID uint64
	cmd   string
	argv  []string
}

// NewSpawner returns a new Spawner that can run maxProcs processes in
// parallel. It kills processes when they exceed a runtime of maxRuntime
// seconds.
func NewSpawner(maxProcs int, maxRuntime uint) Spawner {
	worker := work.NewWorker(maxProcs, func(payload work.Payload) interface{} {
		params := payload.Data.(cmdParams)
		quitChan := process.StartWithTimeout(params.cmd, params.argv, time.Duration(maxRuntime)*time.Second)
		log.Info("Program %s exited with status %d\n", params.cmd, <-quitChan)
		return nil
	}, false)

	return Spawner{
		worker: worker,
	}
}

// Spawn spawns a new process and returns its ID. The returned ID is not the
// PID but an internal ID you can use to keep track of your processes. If an
// error occurs, a non-nil error is returned.
func (s *Spawner) Spawn(cmd string, argv []string) (uint64, error) {
	jobID := s.jobCnt
	err := s.worker.Dispatch(work.Payload{
		cmdParams{
			jobID: jobID,
			cmd:   cmd,
			argv:  argv,
		},
	})
	if err != nil {
		return 0, err
	}
	s.jobCnt += 1
	return jobID, nil
}
