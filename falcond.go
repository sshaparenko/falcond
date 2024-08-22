package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	socket, err := net.Listen("unix", "/tmp/falcon.sock")
	if err != nil {
		log.Fatalf("Failed to listen to a socket: %s\n", err.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/falcon.sock")
		os.Exit(1)
	}()

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection from socket: %s\n", err.Error())
		}

		go func(conn net.Conn) {
			defer conn.Close()

			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("Failed to read from socket: %s\n", err.Error())
				c <- os.Interrupt
			}

			fmt.Printf("Reacieved: %s\n", string(buf[:n]))
		}(conn)
	}
}
