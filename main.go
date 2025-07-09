package main

import (
	"fmt"
	ginserver "redesigned-computing-machine/gin-server"
	"sync"
)

func main() {
	fmt.Println("Hello, redesigned computing machine!")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ginserver.StartServer()
	}()

	wg.Wait()
}
