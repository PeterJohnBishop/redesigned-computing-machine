package main

import (
	"fmt"
	ginserver "redesigned-computing-machine/gin-server"
	"sync"
	"time"
)

func main() {
	fmt.Println("Hello, redesigned computing machine!")
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ginserver.StartServer()
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ginserver.PeerDiscovery()
		}
	}()

	wg.Wait()
}
