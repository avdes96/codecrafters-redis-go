package main

import (
	"flag"
	"log"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/logger"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

func main() {
	logger.InitLogger()
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error opening file %s", err)
	}
	dbdir := flag.String("dir", currentDir, "The directory of the rdb file to initialise the redis cache with.")
	dbfilename := flag.String("dbfilename", "defaultdb", "The rdb file to initialise the redis cache with.")
	port := flag.String("port", "6379", "The port number to initialise the redis cache on.")
	replicaof := flag.String("replicaof", "", "The \"<HOSTNAME> <PORT>\" which this redis cache is a replica of.")
	flag.Parse()
	configParams := make(map[string]string)
	configParams["dir"] = *dbdir
	configParams["dbfilename"] = *dbfilename
	configParams["port"] = *port

	replicationInfo := utils.NewReplicationInfo(*replicaof)
	r, err := server.New(configParams, replicationInfo)
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}
	r.SyncWithMaster()
	r.Run()
}
