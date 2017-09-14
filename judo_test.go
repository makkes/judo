package judo_test

import "github.com/makkes/judo"

func Example() {
	// create a spawner with a pool of 100 goroutines and a maximum runtime per
	// process of 20 seconds.
	spawner := judo.NewSpawner(100, 20)
	quitCh := make(chan struct{})
	// spawn »/bin/sleep 1«
	spawner.Spawn("/bin/sleep", []string{"1"}, quitCh)
	// wait for sleep to quit
	<-quitCh
	// tear down the spawner and its resources
	spawner.Quit()
}
