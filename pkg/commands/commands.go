package commands

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
)

/*
	[13] - Enter
	[127] - Backspace
*/

func RunSync(ttyPath string) {
	fmt.Println(ttyPath)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/falcon.sock")
		os.Exit(1)
	}()

	buff := make([]byte, 32*1024)

	tty, err := os.OpenFile(ttyPath, os.O_RDWR, fs.ModeDevice)
	if err != nil {
		log.Fatalf("Failed to open tty: %s\n", err.Error())
	}

	// syscall.SetNonblock(int(tty.Fd()), true)
	var termios syscall.Termios

	syscall.Syscall(syscall.TCGETS)

	defer tty.Close()

	// ptm, pts, err := pty.Open()
	// if err != nil {
	// 	log.Fatalf("Failed to open PTY: %s\n", err.Error())
	// }

	for {
		//log.Println("reading...")
		n, err := tty.Read(buff)
		if err != nil {
			log.Fatalf("Failed to read from TTY: %s\n", err.Error())
		}
		// log.Println(buff)
		log.Println(buff[:n])
		log.Println(string(buff[:n]))
	}
}

// change io.Copy to castom algorithm for communication between tty and pty
func Run(ttyPath string) {
	fmt.Println(ttyPath)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/falcon.sock")
		fmt.Println()
		os.Exit(1)
	}()

	ptmx, err := os.OpenFile(ttyPath, os.O_RDWR, fs.ModeDevice)
	if err != nil {
		log.Fatalf("Failed to open the tty device: %s\n", err.Error())
	}

	ptm, tty, err := pty.Open()
	if err != nil {
		log.Fatalf("Failed to open the tty device: %s\n", err.Error())
	}

	defer ptm.Close()
	defer tty.Close()
	defer ptmx.Close()

	var wg sync.WaitGroup
	// tty -> ptm
	go func() {
		w, err := io.Copy(ptm, ptmx)
		if err != nil {
			log.Fatalf("Failed to copy output from ptmx to ptm: %s\n", err.Error())
		}
		fmt.Println(w)
	}()

	// ptm -> pts
	go func() {
		if _, err := io.Copy(tty, ptm); err != nil {
			log.Fatalf("Failed to copy from ptm to tty: %s\n", err.Error())
		}
	}()

	// pts -> ptm
	go func() {
		if _, err := io.Copy(ptm, tty); err != nil {
			log.Fatalf("Failed to copy from tty to ptm: %s\n", err.Error())
		}
	}()

	wg.Add(1)
	// ptm -> tty
	go func() {
		w, err := io.Copy(ptmx, ptm)
		if err != nil {
			log.Fatalf("Failed to copy from ptm to ptmx: %s\n", err.Error())
		}
		fmt.Println(w)
		wg.Done()
	}()

	wg.Wait()
}
