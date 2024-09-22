package utils

import (
	"os"
	"path/filepath"
	"strings"
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

func CheckExistingWallet() string {
	walletDir := "wallet"

	// Check if wallet directory exists
	if _, err := os.Stat(walletDir); os.IsNotExist(err) {
		return ""
	}

	// Read wallet directory
	files, err := os.ReadDir(walletDir)
	if err != nil {
		return ""
	}

	// Look for .solXENwallet file
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".solXENwallet") {
			return strings.TrimSuffix(file.Name(), ".solXENwallet")
		}
	}

	return ""
}
