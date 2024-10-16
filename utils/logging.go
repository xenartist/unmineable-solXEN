package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/rivo/tview"
)

type LogMessageFunc func(*tview.TextView, string)

var (
	fileLogger *log.Logger
	logFile    *os.File
	once       sync.Once
)

func initFileLogger() {
	execPath, err := os.Executable()
	if err != nil {
		log.Println("Failed to get executable path:", err)
		return
	}

	logFilePath := filepath.Join(filepath.Dir(execPath), "debug.log")

	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open log file:", err)
		return
	}
	fileLogger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogMessage(logView *tview.TextView, message string) {
	fmt.Fprintf(logView, "%s\n", message)
	logView.ScrollToEnd()
}

func LogToFile(message string) {
	once.Do(initFileLogger)
	if fileLogger != nil {
		fileLogger.Println(message)
	}
}

func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}
