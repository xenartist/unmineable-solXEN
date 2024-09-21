package utils

import (
	"os"
	"path/filepath"
)

var GLOBAL_WORK_DIR string

func XoosInit() {
	var err error
	GLOBAL_WORK_DIR, err = os.Getwd()
	if err != nil {
		// Handle error, can choose to panic or set a default value
		panic("Unable to get current working directory: " + err.Error())
	}
	// Ensure the path is absolute
	GLOBAL_WORK_DIR, err = filepath.Abs(GLOBAL_WORK_DIR)
	if err != nil {
		panic("Unable to get absolute path: " + err.Error())
	}
}
