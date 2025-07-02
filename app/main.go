package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)



func main() {
	r, err := server.New()
	if err != nil {
		fmt.Println("Failed to create server")
		os.Exit(1)
	}
	r.Run()
}
