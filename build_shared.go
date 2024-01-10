//go:build shared

package main

import (
	"C"

	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/elgatito/elementum/exit"
)

var (
	readFile  *os.File
	writeFile *os.File
)

func initShared() {
	exit.Reset()
	exit.IsShared = true
}

func initLog(arg string) {
	logPath = arg

	originalStdout := os.Stdout

	readFile, writeFile, _ = os.Pipe()

	os.Stdout = writeFile
	os.Stderr = writeFile

	go func() {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("Could not open log file '%s' for writing: %s\n", logPath, err)
			return
		}
		defer logFile.Close()

		os.Stdout.WriteString(fmt.Sprintf("Redirecting Stdout/Stderr to %s\r\n", logPath))
		scanner := bufio.NewScanner(readFile)
		for scanner.Scan() {
			s := scanner.Text() + "\r\n"
			logFile.WriteString(s)
			originalStdout.WriteString(s)
		}
	}()
}

//export start
func start() {
	initShared()

	main()
}

//export startWithLog
func startWithLog(log *C.char) int {
	initLog(C.GoString(log))

	main()
	closeFiles()

	return exit.Code
}

//export startWithArgs
func startWithArgs(args *C.char) int {
	initShared()

	exit.Args = C.GoString(args)
	main()

	return exit.Code
}

//export startWithLogAndArgs
func startWithLogAndArgs(logPath, args *C.char) int {
	initShared()
	initLog(C.GoString(logPath))

	exit.Args = C.GoString(args)
	main()
	closeFiles()

	// Give time to write everything to log files
	time.Sleep(1 * time.Second)

	return exit.Code
}

func closeFiles() {
	if readFile != nil {
		readFile.Close()
	}
	if writeFile != nil {
		writeFile.Close()
	}
}
