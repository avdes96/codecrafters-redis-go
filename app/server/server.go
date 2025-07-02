package server

import (
	"fmt"
	"io"
	"net"
	"os"
)

type redisServer struct {
	listener net.Listener
}

func New() (*redisServer, error) {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return nil, err
	}
	return &redisServer{listener: l}, nil
}

func (r *redisServer) Run() error {
	defer r.listener.Close()
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go r.handleConnection(conn)
	}
}

func (r *redisServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer); 
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from connection %s", err.Error())
		}
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection %s", err.Error())
		}
	}
}