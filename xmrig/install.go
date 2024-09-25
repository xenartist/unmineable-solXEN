package xenblocks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"xoon/utils"

	"github.com/rivo/tview"
)

const (
	XMRIG_MINER_DIR  = "xmrigMiner"
	CONFIG_FILE_NAME = "config.txt"
)

// ReadConfigFile reads the content of config.txt
func ReadConfigFile(logView *tview.TextView, logMessage utils.LogMessageFunc) (string, error) {
	configPath := filepath.Join(utils.GLOBAL_WORK_DIR, XMRIG_MINER_DIR, CONFIG_FILE_NAME)
	logMessage(logView, "Debug: ReadConfigFile from "+configPath)
	content, err := os.ReadFile(configPath)
	if err != nil {
		logMessage(logView, "Error reading config file: "+err.Error())
		return "", err
	}
	logMessage(logView, "Config file read successfully")
	return string(content), nil
}

// WriteConfigFile writes or updates the content of config.txt
func WriteConfigFile(content string, logView *tview.TextView, logMessage utils.LogMessageFunc) error {
	configPath := filepath.Join(utils.GLOBAL_WORK_DIR, XMRIG_MINER_DIR, CONFIG_FILE_NAME)
	var err = os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		logMessage(logView, "Error writing config file: "+err.Error())
		return err
	}
	logMessage(logView, "Config file written successfully")
	return nil
}

// CreateXenblocksMinerDir creates the xenblocksMiner directory and config file if they don't exist
func CreateXmrigMinerDir(logView *tview.TextView, logMessage utils.LogMessageFunc) error {
	xmrigMinerPath := filepath.Join(utils.GLOBAL_WORK_DIR, XMRIG_MINER_DIR)
	if _, err := os.Stat(xmrigMinerPath); os.IsNotExist(err) {
		err = os.Mkdir(xmrigMinerPath, 0755)
		if err != nil {
			logMessage(logView, fmt.Sprintf("Error creating xenblocksMiner directory: %v", err))
			return err
		}
		logMessage(logView, "xenblocksMiner directory created successfully")
	}

	configPath := filepath.Join(xmrigMinerPath, CONFIG_FILE_NAME)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		content := "account_address=ETH address with uppercase and lowercase\ndevfee_permillage=2"
		err = os.WriteFile(configPath, []byte(content), 0644)
		if err != nil {
			logMessage(logView, fmt.Sprintf("Error creating config file: %v", err))
			return err
		}
		logMessage(logView, "Config file created successfully")
	}

	return nil
}

// InstallXENBLOCKS handles the installation of XENBLOCKS
func InstallXmrig(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc) {

	// Log the start of the installation process
	if isXmrigInstalled(logView, logMessage) {
		return
	}

	logMessage(logView, "Starting XenblocksMiner installation...")

	downloadedPath, err := downloadXmrig(logView, logMessage)
	if err != nil {
		return
	}

	// Extract XenblocksMiner
	xenblocksMinerPath, err := extractXmrig(logView, logMessage, downloadedPath)
	if err != nil {
		return
	}

	logMessage(logView, fmt.Sprintf("XenblocksMiner installed successfully at: %s", xenblocksMinerPath))
}

func isXmrigInstalled(logView *tview.TextView, logMessage utils.LogMessageFunc) bool {
	var executableName string
	if runtime.GOOS == "windows" {
		executableName = "xmrig.exe"
	} else {
		executableName = "xmrig"
	}

	executablePath := filepath.Join(utils.GLOBAL_WORK_DIR, XMRIG_MINER_DIR, executableName)
	if _, err := os.Stat(executablePath); err == nil {
		logMessage(logView, "XenblocksMiner is already installed. No need to install again.")
		return true
	}
	return false
}

func downloadXmrig(logView *tview.TextView, logMessage utils.LogMessageFunc) (string, error) {
	var url, fileName string

	if runtime.GOOS == "windows" {
		url = "https://github.com/xmrig/xmrig/releases/download/v6.22.0/xmrig-6.22.0-msvc-win64.zip"
		fileName = "xmrig-6.22.0-msvc-win64.zip"
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "amd64" {
			url = "https://github.com/xmrig/xmrig/releases/download/v6.22.0/xmrig-6.22.0-macos-x64.tar.gz"
			fileName = "xmrig-6.22.0-macos-x64.tar.gz"
		} else if runtime.GOARCH == "arm64" {
			url = "https://github.com/xmrig/xmrig/releases/download/v6.22.0/xmrig-6.22.0-macos-arm64.tar.gz"
			fileName = "xmrig-6.22.0-macos-arm64.tar.gz"
		}
	} else {
		url = "https://github.com/xmrig/xmrig/releases/download/v6.22.0/xmrig-6.22.0-linux-static-x64.tar.gz"
		fileName = "xmrig-6.22.0-linux-static-x64.tar.gz"
	}

	// Construct the full file path
	filePath := filepath.Join(utils.GLOBAL_WORK_DIR, XMRIG_MINER_DIR, fileName)

	// Prepare the curl command
	cmd := exec.Command("curl", "-L", "-o", filePath, url)

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logMessage(logView, fmt.Sprintf("Error downloading file: %v\nOutput: %s", err, string(output)))
		return "", err
	}

	logMessage(logView, fmt.Sprintf("File downloaded successfully to: %s", filePath))
	return filePath, nil
}

func extractXmrig(logView *tview.TextView, logMessage utils.LogMessageFunc, downloadedPath string) (string, error) {
	// Get the directory of the downloaded file
	dir := filepath.Dir(downloadedPath)

	// Extract the tar.gz file
	cmd := exec.Command("tar", "-zxvf", downloadedPath, "-C", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logMessage(logView, fmt.Sprintf("Error extracting file: %v\nOutput: %s", err, string(output)))
		return "", err
	}
	logMessage(logView, "File extracted successfully")

	var executablePath string
	// Make the file executable
	if runtime.GOOS != "windows" {
		// Construct the path to the extracted executable
		executablePath = filepath.Join(dir, "/xmrig-6.22.0/xmrig")
		// For Linux and other Unix-like systems
		cmd = exec.Command("chmod", "+x", executablePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			logMessage(logView, fmt.Sprintf("Error making file executable: %v\nOutput: %s", err, string(output)))
			return "", err
		}
		logMessage(logView, "File permissions updated successfully")
	} else {
		// Construct the path to the extracted executable
		executablePath = filepath.Join(dir, "\\xmrig-6.22.0\\xmrig.exe")
		// For Windows, no need to change permissions
		logMessage(logView, "File permissions update not required on Windows")
	}

	// Return the path to the executable
	return executablePath, nil
}
