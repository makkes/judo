# Judo - your robust process forker

This is an experiment I'm conducting that involves forking processes in an
efficient way. When the main process is killed, all child processes shall also
be killed. Also, processes that take too long shall be killed after a specified
time. So, this is more or less a robust and efficient process forker. 

Judo provides both a command-line program as well as a library for developing
custom programs that need to fork processes.

## Get it

```sh
go get github.com/makkes/judo/...
```

## Run it

```sh
judo
```

Now that Judo is running, just type in a command that you'd like to be invoked
(e.g. `/bin/ls`) and you'll see its output on the console. The default maximum
runtime for a program is 60 seconds. So when you type `/bin/sleep 65` you'll see
that Judo kills the process after one minute.

## Develop custom programs

See the [GoDoc](https://godoc.org/github.com/makkes/judo).

## Understand it

### Process management

Judo uses the [work](https://godoc.org/github.com/justsocialapps/justlib/work)
package from [Just Social](https://github.com/justsocialapps/) to manage
sub-processes, i.e. it starts with a specific number of goroutines and
dispatches work to them via channels. This ensures that the number of processes
forked by Judo doesn't get out of hand.

### Child process control

Judo sets the PGID of every child process to the PID of Judo itself, so you can
easily kill all child processes by executing `kill -PID`.

Also, Judo sets the parent process death signal of every child process to
SIGTERM. This means that when Judo is killed (e.g. using Ctrl+C or by sending
the HUB or KILL signal to Judo) then all child processes will receive the TERM
signal and eventually be killed. This ensures that we leave no resources alive
after Judo dies.
