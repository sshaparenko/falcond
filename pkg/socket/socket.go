package socket

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sshaparenkos/falcond/pb"
	"github.com/sshaparenkos/falcond/pkg/commands"
	"google.golang.org/protobuf/proto"
)

var messages = make(chan []byte, 1024)

func Listen(network string, address string, wg *sync.WaitGroup) {
	socket, err := net.Listen("unix", "/tmp/falcon.sock")
	if err != nil {
		log.Fatalf("Failed to listen to a socket: %s\n", err.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		wg.Done()
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

			buff := make([]byte, 4096)
			n, err := conn.Read(buff)
			if err != nil {
				fmt.Printf("Failed to read from socket: %s\n", err.Error())
				c <- os.Interrupt
			}

			messages <- buff[:n]
		}(conn)
	}
}

func Operate(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		for {
			command := <-messages
			message := &pb.Feather{}
			if err := proto.Unmarshal(command, message); err != nil {
				log.Fatalf("Failed to unmarshal proto message: %s\n", err.Error())
			}
			switch message.Command {
			case "run":
				commands.RunSync(message.TtyPath)
				// commands.Run(message.TtyPath)
			default:

			}
		}
	}()
}
