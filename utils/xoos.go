package utils

import (
	"os"
	"path/filepath"
)

// var GLOBAL_WORK_DIR string

func GetExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

func XoosInit() {
	// GLOBAL_WORK_DIR = getExecutablePath()
	initFileLogger()
}
