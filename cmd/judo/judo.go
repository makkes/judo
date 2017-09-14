package main

import (
	"bufio"
	"flag"
	"os"
	"strings"

	golog "log"

	"github.com/makkes/judo"
	log "gopkg.in/justsocialapps/justlib.v1/logging"
)

func main() {
	log.SetLevel(log.WARN)
	golog.SetFlags(golog.LstdFlags | golog.Lshortfile)

	loglevel := flag.String("l", "WARN", "The loglevel to use. One of DEBUG, INFO, WARN, ERROR")
	maxRuntime := flag.Uint("r", 60, "The maximum time that a child process is allowed to run before it is killed")
	maxProcs := flag.Int("p", 100, "The maximum number of parallel executing processes")
	flag.Parse()

	level, err := log.LevelFromString(*loglevel)
	if err != nil {
		log.Fatal("Unknown loglevel: %s", *loglevel)
	}
	log.SetLevel(*level)

	spawner := judo.NewSpawner(*maxProcs, *maxRuntime)

	reader := bufio.NewReader(os.Stdin)
	for {
		line, _, _ := reader.ReadLine()
		if len(line) == 0 {
			continue
		}
		cmdArr := strings.Split(string(line), " ")
		var argv []string
		if len(cmdArr) > 1 {
			argv = cmdArr[1:]
		}

		quitChan := make(chan struct{})
		err := spawner.Spawn(cmdArr[0], argv, quitChan)
		if err != nil {
			log.Error("Error starting job: %s", err)
		} else {
			log.Info("Job started")
			go func(cmd string, quitChan chan struct{}) {
				<-quitChan
				log.Info("Job %s ended", cmd)
			}(cmdArr[0], quitChan)
		}
	}
}
