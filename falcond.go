package main

import (
	"sync"

	"github.com/sshaparenkos/falcond/pkg/socket"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go socket.Listen("unix", "/tmp/falcon.sock", &wg)
	wg.Add(1)
	go socket.Operate(&wg)
	wg.Wait()
}
