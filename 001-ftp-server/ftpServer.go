package main

import (
	"bufio"
	"fmt"
	"net"
)

var logFile = "/go/src/myftp/MyFtpLog.log"

// FtpServerSettings ...
type FtpServerSettings struct {
	listenAddr string
	listenPort int
}

// FtpServer ...
type FtpServer struct {
	logger   *FtpLogger
	settings *FtpServerSettings
	listener *net.Listener
}

// CreateFtpServer ...
func CreateFtpServer(ip string, port int) (*FtpServer, error) {
	ftpServer := &(FtpServer{nil, nil, nil})
	var err error
	ftpServer.logger, err = CreateFtpLogger(logFile)
	if err != nil {
		return nil, err
	}
	ftpServer.settings = &(FtpServerSettings{ip + ":" + fmt.Sprintf("%v", port), port})
	ftpServer.logger.Log("Create a FTP server.")
	return ftpServer, nil
}

// Listen ...
func (ftpServer *FtpServer) Listen() error {
	listener, err := net.Listen("tcp", ftpServer.settings.listenAddr)
	if err != nil {
		ftpServer.logger.Log("Cannot start listener!")
		return err
	}
	ftpServer.listener = &listener
	ftpServer.logger.Log("FTP server starts to listen.")
	return nil
}

// Serve ...
func (ftpServer *FtpServer) Serve() {
	for {
		conn, err := (*ftpServer.listener).Accept()
		if err != nil {
			ftpServer.logger.Log("FTP server accepts error!")
			break
		}
		go ftpServer.handleClient(conn)
		ftpServer.logger.Log("FTP server accepts a client.")
	}
}

func (ftpServer *FtpServer) handleClient(conn net.Conn) {
	pi, err := CreateFtpPI(conn, ftpServer.logger)
	defer conn.Close()
	if err != nil {
		tmpWriter := bufio.NewWriter(conn)
		tmpWriter.Write([]byte(fmt.Sprintf("500 Server Internal Error %s\r\n", err.Error())))
		tmpWriter.Flush()
		return
	}
	pi.Serve()
}
