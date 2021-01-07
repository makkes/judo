// Package process implements the low-level process handling for judo. It
// contains routines for starting a single process.
package process

import (
	"os"
	"syscall"
	"time"

	log "gopkg.in/justsocialapps/justlib.v1/logging"
)

// StartWithTimeout works just like Start with the addition that it kills the
// command as soon as the given duration expires. The returned channel signals
// the exit code of the command.
func StartWithTimeout(cmd string, argv []string, timeout time.Duration) <-chan int {
	killChan, internalQuitChan := Start(cmd, argv, "")
	quitChan := make(chan int)

	go func() {
		timer := time.NewTimer(timeout)
		select {
		case <-timer.C:
			killChan <- struct{}{}
			quitChan <- -1
		case exitStatus := <-internalQuitChan:
			timer.Stop()
			quitChan <- exitStatus.ExitCode
		}
	}()

	return quitChan
}

type Result struct {
	ExitCode int
	Err      error
}

// Start executes the given command with the given parameters and returns two
// channels: The first channel is used to kill the executed command and all
// sub-processes immediately; just send an empty struct to it. The second
// channel signals the exit code of the command as soon as it is quit. This
// code is set to -1 when the kill channel is used to kill the command.
func Start(cmd string, argv []string, dir string) (chan<- struct{}, <-chan Result) {
	killChan := make(chan struct{})
	quitChan := make(chan Result)

	// start the process in a goroutine and quit the goroutine when the
	// process exits.
	go func() {
		proc, err := os.StartProcess(cmd, append([]string{cmd}, argv...), &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Dir:   dir,
			Sys: &syscall.SysProcAttr{
				// set the PGID of the child process so we can kill
				// all children with kill -PGID.
				Setpgid: true,
				// When the parent dies, send the TERM signal
				// to the child.
				Pdeathsig: syscall.SIGTERM,
			}})
		if err != nil {
			quitChan <- Result{
				ExitCode: -1,
				Err: err,
			}
			return
		}

		killRoutineQuitChan := make(chan struct{})
		mainRoutineQuitChan := make(chan struct{})
		// this goroutine waits for a command to kill the process.
		go func() {
			select {
			case <-killChan:
				err := syscall.Kill(-proc.Pid, syscall.SIGKILL)
				if err != nil {
					if syscall.ESRCH != err { // there was a real error killing the process
						log.Error("Error killing process: %s\n", err)
					}
				}
				mainRoutineQuitChan <- struct{}{} // signal the main routine to quit
			case <-killRoutineQuitChan:
				// the main routine wants me to quit
			}
			close(killChan)
		}()

		// wait for the program to exit (or be killed)
		state, err := proc.Wait()
		exitStatus := state.Sys().(syscall.WaitStatus).ExitStatus()

		// there are 2 possible outcomes here:
		//
		// 1. the program exited by itself with an exit status. Then we
		//    pass that exit status to the quitChan.
		// 2. the program has been killed by our kill routine (above).
		//    In this case we receive a signal via the
		//    mainRoutineQuitChan and just exit.
		select {
		case quitChan <- Result{
			ExitCode: exitStatus,
		}:
		case <-mainRoutineQuitChan:
		}

		// close the channel so that our kill routine quits.
		close(killRoutineQuitChan)
	}()
	return killChan, quitChan
}
