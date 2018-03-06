package main

import (
	"log"
	"os"
)

// FtpLogger The logger used for ftp server
type FtpLogger struct {
	LogFileName string
	logger      *log.Logger
	logFile     *os.File
}

// CreateFtpLogger ...
func CreateFtpLogger(filename string) (*FtpLogger, error) {
	ftpLogger := &(FtpLogger{"", nil, nil})
	ftpLogger.LogFileName = filename
	var err error
	ftpLogger.logFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	ftpLogger.logger = log.New(ftpLogger.logFile, "FtpLogger: ", log.Lshortfile)
	return ftpLogger, nil
}

// Log ...
func (ftpLogger *FtpLogger) Log(content string) {
	ftpLogger.logger.Print(content)
}
