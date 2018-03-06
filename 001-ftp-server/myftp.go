package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
)

const (
	minPort = 2121
	maxPort = 2200
)

func main() {
	portV := flag.Int("p", 2121, "listening port")
	hostV := flag.String("a", "", "binding address")
	dirV := flag.String("d", RootDir, "change current directory")
	nativeV := flag.Int("native", 0, "run in native system")

	flag.Parse()

	RootDir = *dirV
	if *portV < minPort || *portV > maxPort {
		fmt.Printf("Required port number in [%v, %v]\n", minPort, maxPort)
		os.Exit(1)
	}
	fmt.Printf("Port: %v, Host: %v, Directory: %v\n", *portV, *hostV, *dirV)

	if *nativeV == 1 {
		RootDir = "./ftpdir"
		AccountFile = "./ftpAccounts.dat"
		logFile = "./MyFtpLog.log"
	}

	go handleSignal()

	server, err := CreateFtpServer(*hostV, *portV)
	if err != nil {
		fmt.Println("Cannot create server!")
	}
	server.Listen()
	server.Serve()
}

func handleSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	for {
		switch s := <-ch; s {
		case os.Interrupt:
			fmt.Println("SIGTERM Signal!")
			os.Exit(0)
		}
	}
}
