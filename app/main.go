package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	dbdir := flag.String("dir", currentDir, "The directory of the rdb file to initialise the redis cache with.")
	dbfilename := flag.String("dbfilename", "defaultdb", "The rdb file to initialise the redis cache with.")
	port := flag.String("port", "6379", "The port number to initialise the redis cache on.")
	flag.Parse()
	configParams := make(map[string]string)
	configParams["dir"] = *dbdir
	configParams["dbfilename"] = *dbfilename
	configParams["port"] = *port
	r, err := server.New(configParams)
	if err != nil {
		fmt.Println("Failed to create server")
		os.Exit(1)
	}
	r.Run()
}
