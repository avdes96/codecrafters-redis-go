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
	dbdir := flag.String("dir", currentDir, "CONFIG GET --dir <directory>")
	dbfilename := flag.String("dbfilename", "defaultdb", "CONFIG GET --dir <directory>")
	flag.Parse()
	configParams := make(map[string]string)
	configParams["dir"] = *dbdir
	configParams["dbfilename"] = *dbfilename
	r, err := server.New(configParams)
	if err != nil {
		fmt.Println("Failed to create server")
		os.Exit(1)
	}
	r.Run()
}
