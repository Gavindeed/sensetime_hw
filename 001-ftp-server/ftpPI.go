package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

// RootDir ...
var RootDir = "/go/src/myftp/ftpdir"

// AccountFile ...
var AccountFile = "/go/src/myftp/ftpAccounts.dat"

// ReplyMap ...
var ReplyMap = map[int]string{
	200: "Command okay.",
	500: "Syntax error, command unrecognized.\r\nThis may include errors such as command line too long.",
	501: "Syntax error in parameters or arguments.",
	202: "Command not implemented, superfluous at this site.",
	502: "Command not implemented.",
	503: "Bad sequence of commands.",
	504: "Command not implemented for that parameter.",
	110: "Restart marker reply.\r\nIn this case, the text is exact and not left to the particular implementation; it must read: \r\nMARK yyyy = mmmm\r\nWhere yyyy is User-process data stream marker, and mmmm	server's equivalent marker (note the spaces between markers and \"=\").",
	211: "System status, or system help reply.",
	212: "Directory status.",
	213: "File status.",
	214: "Help message.\r\nOn how to use the server or the meaning of a particular non-standard command.  This reply is useful only to the human user.",
	215: "NAME system type.\r\nWhere NAME is an official system name from the list in the Assigned Numbers document.",
	120: "Service ready in nnn minutes.",
	220: "Service ready for new user.",
	221: "Service closing control connection.\r\nLogged out if appropriate.",
	421: "Service not available, closing control connection.\r\nThis may be a reply to any command if the service knows it must shut down.",
	125: "Data connection already open; transfer starting.",
	225: "Data connection open; no transfer in progress.",
	425: "Can't open data connection.",
	226: "Closing data connection.\r\nRequested file action successful (for example, file transfer or file abort).",
	426: "Connection closed; transfer aborted.",
	227: "Entering Passive Mode (h1,h2,h3,h4,p1,p2).",
	230: "User logged in, proceed.",
	530: "Not logged in.",
	331: "User name okay, need password.",
	332: "Need account for login.",
	532: "Need account for storing files.",
	150: "File status okay; about to open data connection.",
	250: "Requested file action okay, completed.",
	257: "\"PATHNAME\" created.",
	350: "Requested file action pending further information.",
	450: "Requested file action not taken.\r\nFile unavailable (e.g., file busy).",
	550: "Requested action not taken.\r\nFile unavailable (e.g., file not found, no access).",
	451: "Requested action aborted. Local error in processing.",
	551: "Requested action aborted. Page type unknown.",
	452: "Requested action not taken.\r\nInsufficient storage space in system.",
	552: "Requested file action aborted.\r\nExceeded storage allocation (for current directory or dataset).",
	553: "Requested action not taken.\r\nFile name not allowed.",
}

const (
	// TypeBinary ...
	TypeBinary = 0
	// TypeASCII ...
	TypeASCII = 1
)

// FtpPI ...
type FtpPI struct {
	conn     net.Conn
	user     string
	pass     string
	auth     bool
	comm     string
	para     string
	curPath  string
	dtp      *FtpDTP
	accounts []Account
	logger   *FtpLogger
	writer   *bufio.Writer
	reader   *bufio.Reader
	typeT    int
}

// CreateFtpPI ...
func CreateFtpPI(conn net.Conn, logger *FtpLogger) (*FtpPI, error) {
	dtp, err := CreateFtpDTP()
	if err != nil {
		logger.Log("Cannot create DTP!")
		return nil, err
	}
	pi := &(FtpPI{conn, "", "", false, "", "", RootDir, dtp, make([]Account, 0), logger, nil, nil, 0})
	pi.accounts, err = CreateAccountListFromFile(AccountFile)
	if err != nil {
		logger.Log("Cannot create account list!")
		return nil, err
	}
	pi.writer = bufio.NewWriter(conn)
	pi.reader = bufio.NewReader(conn)
	return pi, nil
}

// Serve ...
func (ftpPI *FtpPI) Serve() {
	// reader := bufio.NewReader(os.Stdin)
	msg, err := ftpPI.welcome()
	if err == nil {
		ftpPI.writeMsg(220, msg)
		ftpPI.logger.Log("The server begins to serve.")
	} else {
		ftpPI.writeMsgCode(500)
		ftpPI.logger.Log("The server cannot serve.")
		return
	}
	for {
		ins, err := ftpPI.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				ftpPI.conn.Close()
				ftpPI.logger.Log("Remote client stop connection!")
				return
			}
			ftpPI.logger.Log("Read command error!")
			continue
		}
		// inses := strings.Fields(ins)
		inses := strings.SplitN(strings.Trim(ins, "\r\n"), " ", 2)
		ftpPI.comm = ""
		ftpPI.para = ""
		if len(inses) == 1 {
			ftpPI.comm = inses[0]
			ftpPI.para = ""
		} else {
			ftpPI.comm = inses[0]
			ftpPI.para = inses[1]
		}
		if ftpPI.comm != "" {
			quit, _ := ftpPI.HandleCommand()
			if quit {
				return
			}
		}
	}
}

func (ftpPI *FtpPI) writeLine(content string) {
	ftpPI.writer.Write([]byte(content))
	ftpPI.writer.Write([]byte("\r\n"))
	ftpPI.writer.Flush()
}

func (ftpPI *FtpPI) writeMsg(code int, content string) {
	ftpPI.writeLine(fmt.Sprintf("%v %v", code, content))
}

func (ftpPI *FtpPI) writeMsgCode(code int) {
	ftpPI.writeLine(fmt.Sprintf("%v %v", code, ReplyMap[code]))
}

func (ftpPI *FtpPI) welcome() (string, error) {
	return fmt.Sprintf("Welcome to MyFTP, your user name is %v, your id is %v, your current working directory is %v, your ip address is %v", ftpPI.user, 1, ftpPI.curPath, ftpPI.conn.RemoteAddr()), nil
}

// HandleCommand ...
func (ftpPI *FtpPI) HandleCommand() (bool, error) {
	switch ftpPI.comm {
	case "USER":
		return false, ftpPI.HandleUSER()
	case "PASS":
		return false, ftpPI.HandlePASS()
	case "SYST":
		return false, ftpPI.HandleSYST()
	case "FEAT":
		return false, ftpPI.HandleFEAT()
	case "TYPE":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandleTYPE()
	case "PASV":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandlePASV()
	case "LIST":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandleLIST()
	case "CWD":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandleCWD()
	case "CDUP":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		ftpPI.para = ".."
		return false, ftpPI.HandleCWD()
	case "PWD":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandlePWD()
	case "RETR":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandleRETR()
	case "STOR":
		if !ftpPI.auth {
			ftpPI.writeMsgCode(530)
			return false, fmt.Errorf("user not log in")
		}
		return false, ftpPI.HandleSTOR()
	case "QUIT":
		return true, ftpPI.HandleQUIT()
	default:
		ftpPI.writeMsgCode(502)
		ftpPI.logger.Log(fmt.Sprintf("Command %v is not supported!", ftpPI.comm))
		return false, fmt.Errorf("command %v not supported", ftpPI.comm)
	}
}

// HandleUSER ...
func (ftpPI *FtpPI) HandleUSER() error {
	ftpPI.user = ftpPI.para
	if ftpPI.user == "" {
		ftpPI.writeMsgCode(332)
		return fmt.Errorf("invalid user name")
	}
	ftpPI.auth = false
	// fmt.Println("Please enter the password for user", ftpPI.user)
	ftpPI.writeMsgCode(331)
	return nil
}

// HandlePASS ...
func (ftpPI *FtpPI) HandlePASS() error {
	ftpPI.pass = ftpPI.para
	_, err := Authenticate(ftpPI.user, ftpPI.pass, ftpPI.accounts)
	if err != nil {
		ftpPI.logger.Log("Username or password wrong!")
		ftpPI.writeMsgCode(530)
		return err
	}
	// ftpPI.curPath = RootDir + dir
	ftpPI.curPath = RootDir
	ftpPI.dtp.userRootPath = ftpPI.curPath
	ftpPI.auth = true
	ftpPI.logger.Log(fmt.Sprintf("User %v logged in, Dir: %v", ftpPI.user, ftpPI.curPath))
	// fmt.Println("User", ftpPI.user, "log in!")
	ftpPI.writeMsgCode(230)
	return nil
}

// HandleSYST ...
func (ftpPI *FtpPI) HandleSYST() error {
	ftpPI.writeMsg(215, "Type: Unix")
	return nil
}

// HandleFEAT ...
func (ftpPI *FtpPI) HandleFEAT() error {
	ftpPI.writeMsg(211, "I do not know how to describe my features...")
	return nil
}

// HandleTYPE ...
func (ftpPI *FtpPI) HandleTYPE() error {
	switch ftpPI.para {
	case "I":
		ftpPI.typeT = TypeBinary
		ftpPI.writeMsg(200, "Set type to binary.")
	case "A":
		ftpPI.typeT = TypeASCII
		ftpPI.writeMsg(200, "Set type to ASCII.")
	default:
		ftpPI.writeMsgCode(504)
		return fmt.Errorf("parameter not supported")
	}
	return nil
}

// HandlePASV ...
func (ftpPI *FtpPI) HandlePASV() error {
	err := ftpPI.dtp.SetPassive()
	if err != nil {
		ftpPI.writeMsg(451, "Local error in setting passive mode")
		ftpPI.logger.Log("Cannot set passive mode!")
	} else {
		p1 := ftpPI.dtp.transfer.GetPort() / 256
		p2 := ftpPI.dtp.transfer.GetPort() - 256*p1
		ip := strings.Split(ftpPI.conn.LocalAddr().String(), ":")[0]
		ipList := strings.Split(ip, ".")
		ftpPI.writeMsg(227, fmt.Sprintf("Entering Passive Mode (%v,%v,%v,%v,%v,%v).", ipList[0], ipList[1], ipList[2], ipList[3], p1, p2))
	}
	return err
}

// HandleLIST ...
func (ftpPI *FtpPI) HandleLIST() error {
	var path string
	if ftpPI.para == "" {
		path = ftpPI.curPath
	} else if ftpPI.para[0] == '/' {
		path = ftpPI.dtp.userRootPath + ftpPI.para
	} else {
		path = ftpPI.curPath + "/" + ftpPI.para
	}
	if !ftpPI.dtp.ValidPath(path) {
		// fmt.Println("Invalid Path!")
		ftpPI.writeMsgCode(450)
		return fmt.Errorf("invalid path %v", path)
	}
	ftpPI.writeMsg(150, "Opening ASCII mode data connection for file list")
	err := ftpPI.dtp.ListFileInfo(path)
	if err != nil {
		ftpPI.writeMsgCode(451)
		ftpPI.logger.Log("Cannot get the file!")
		return err
	}
	ftpPI.writeMsg(226, "Transfer complete.")
	return nil
}

// HandleCWD ...
func (ftpPI *FtpPI) HandleCWD() error {
	var path string
	if ftpPI.para == "" {
		path = ftpPI.curPath
	} else if ftpPI.para[0] == '/' {
		path = ftpPI.dtp.userRootPath + ftpPI.para
	} else {
		path = ftpPI.curPath + "/" + ftpPI.para
	}
	if !ftpPI.dtp.ValidPath(path) {
		fmt.Println("Invalid Path ", path)
		ftpPI.writeMsgCode(450)
		return fmt.Errorf("invalid path %v", path)
	}
	var err error
	ftpPI.curPath, err = ftpPI.dtp.AbsPath(path)
	if err != nil {
		ftpPI.writeMsgCode(450)
		fmt.Println("Cannot get this path!")
		return err
	}
	newPath := ftpPI.curPath[len(ftpPI.dtp.userRootPath):]
	if newPath == "" {
		newPath = "/"
	}
	fmt.Println("Change current path to", newPath)
	ftpPI.writeMsg(250, "CD worked on "+newPath)
	return nil
}

// HandlePWD ...
func (ftpPI *FtpPI) HandlePWD() error {
	if !ftpPI.dtp.ValidPath(ftpPI.curPath) {
		fmt.Println("Current working path is not valid!")
		return fmt.Errorf("invalid working path %v", ftpPI.curPath)
	}
	path := ftpPI.curPath[len(ftpPI.dtp.userRootPath):]
	if path == "" {
		path = "/"
	}
	// fmt.Println(path)
	ftpPI.writeMsg(257, "\""+path+"\" is the current directory")
	return nil
}

// HandleRETR ...
func (ftpPI *FtpPI) HandleRETR() error {
	var path string
	if ftpPI.para == "" {
		path = ftpPI.curPath
	} else if ftpPI.para[0] == '/' {
		path = ftpPI.dtp.userRootPath + ftpPI.para
	} else {
		path = ftpPI.curPath + "/" + ftpPI.para
	}
	if !ftpPI.dtp.ValidPath(path) {
		fmt.Println("Invalid Path ", path)
		ftpPI.writeMsgCode(450)
		return fmt.Errorf("invalid path %v", path)
	}
	ftpPI.writeMsgCode(150)
	err := ftpPI.dtp.SendFile(path)
	if err != nil {
		ftpPI.writeMsgCode(451)
		ftpPI.logger.Log("Cannot get the file!")
		return err
	}
	ftpPI.writeMsg(226, "Transfer complete.")
	return nil
}

// HandleSTOR ...
func (ftpPI *FtpPI) HandleSTOR() error {
	var path string
	if ftpPI.para == "" {
		path = ftpPI.curPath
	} else if ftpPI.para[0] == '/' {
		path = ftpPI.dtp.userRootPath + ftpPI.para
	} else {
		path = ftpPI.curPath + "/" + ftpPI.para
	}
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		fmt.Println("Invalid Path ", path)
		ftpPI.writeMsgCode(450)
		return fmt.Errorf("invalid path %v", path)
	}
	fatherPath := path[:idx]
	if !ftpPI.dtp.ValidPath(fatherPath) || !ftpPI.dtp.IsDir(fatherPath) {
		fmt.Println("Invalid Path ", path)
		ftpPI.writeMsgCode(450)
		return fmt.Errorf("invalid path %v", path)
	}
	ftpPI.writeMsgCode(150)
	err := ftpPI.dtp.ReceiveFile(path)
	if err != nil {
		ftpPI.writeMsgCode(451)
		ftpPI.logger.Log("Cannot get the file!")
		return err
	}
	ftpPI.writeMsg(226, "Transfer complete.")
	return nil
}

// HandleQUIT ...
func (ftpPI *FtpPI) HandleQUIT() error {
	ftpPI.writeMsg(221, "Goodbye")
	ftpPI.conn.Close()
	return nil
}
