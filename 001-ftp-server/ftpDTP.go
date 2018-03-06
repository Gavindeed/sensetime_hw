package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FtpDTP ...
type FtpDTP struct {
	userRootPath string
	transfer     Transfer
}

// CreateFtpDTP ...
func CreateFtpDTP() (*FtpDTP, error) {
	return &(FtpDTP{"", nil}), nil
}

// AbsPath ...
func (ftpDTP *FtpDTP) AbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// IsDir ...
func (ftpDTP *FtpDTP) IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// ValidPath ...
func (ftpDTP *FtpDTP) ValidPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	if strings.HasPrefix(absPath, ftpDTP.userRootPath) {
		_, err = os.Stat(path)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

const (
	dateFormatYear  = "Jan _2  2006"
	dateFormatTime  = "Jan _2 15:04"
	dateFormatBound = time.Hour * 24 * 30 * 6
	dateFormatMLSD  = "20060102150405"
)

// GetFileInfoString ...
func (ftpDTP *FtpDTP) GetFileInfoString(file os.FileInfo) string {
	modTime := file.ModTime()
	var dateFormat string
	if time.Now().Sub(modTime) > dateFormatBound {
		dateFormat = dateFormatYear
	} else {
		dateFormat = dateFormatTime
	}

	return fmt.Sprintf("%s 1 ftp ftp %12d %s %s", file.Mode(), file.Size(), file.ModTime().Format(dateFormat), file.Name())
}

// ListFileInfo ...
func (ftpDTP *FtpDTP) ListFileInfo(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	var files []os.FileInfo
	if fileInfo.IsDir() {
		files, err = ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		// for _, f := range files {
		// 	fmt.Println(f.Name())
		// }
	} else {
		files = append(files, fileInfo)
		// fmt.Println("Filename: ", fileInfo.Name())
		// fmt.Println("Size: ", fileInfo.Size())
		// fmt.Println("Mode: ", fileInfo.Mode())
		// fmt.Println("Modification Time: ", fileInfo.ModTime())
	}
	conn, err := ftpDTP.transfer.Open()
	if err != nil {
		return err
	}
	fmt.Println("Transfer Open!")
	defer ftpDTP.transfer.Close()
	for _, file := range files {
		fmt.Fprintf(conn, "%s\r\n", ftpDTP.GetFileInfoString(file))
	}
	// conn.Close()
	// ftpDTP.transfer.Close()
	return nil
}

// SetPassive ...
func (ftpDTP *FtpDTP) SetPassive() error {
	var err error
	// if ftpDTP.transfer == nil {
	ftpDTP.transfer, err = CreatePassiveTransfer()
	// }
	if err != nil {
		return err
	}
	return nil
}

// SendFile ...
func (ftpDTP *FtpDTP) SendFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	conn, err := ftpDTP.transfer.Open()
	if err != nil {
		return err
	}
	// defer conn.Close()
	defer ftpDTP.transfer.Close()
	_, err = io.Copy(conn, file)
	if err != nil && err != io.EOF {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

// ReceiveFile ...
func (ftpDTP *FtpDTP) ReceiveFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		// fmt.Println("error1")
		return err
	}
	conn, err := ftpDTP.transfer.Open()
	if err != nil {
		// fmt.Println("error2")
		return err
	}
	// defer conn.Close()
	defer ftpDTP.transfer.Close()
	_, err = io.Copy(file, conn)
	if err != nil && err != io.EOF {
		// fmt.Println("error3", err.Error())
		return err
	}
	err = file.Close()
	if err != nil {
		// fmt.Println("error4")
		return err
	}
	return nil
}
