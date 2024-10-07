package xenblocks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"xoon/utils"

	"github.com/rivo/tview"
)

const (
	SRB_MINER_DIR = "srbMiner"
)

func CreateSrbMinerDir(logView *tview.TextView, logMessage utils.LogMessageFunc) error {
	srbMinerPath := filepath.Join(utils.GetExecutablePath(), SRB_MINER_DIR)
	if _, err := os.Stat(srbMinerPath); os.IsNotExist(err) {
		err = os.Mkdir(srbMinerPath, 0755)
		if err != nil {
			logMessage(logView, fmt.Sprintf("Error creating srbMiner directory: %v", err))
			return err
		}
		logMessage(logView, "srbMiner directory created successfully")
	}

	return nil
}

func InstallSrbMiner(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc) {

	// Log the start of the installation process
	if isSrbMinerInstalled(logView, logMessage) {
		return
	}

	logMessage(logView, "Starting srbMiner installation...")

	downloadedPath, err := downloadSrbMiner(logView, logMessage)
	if err != nil {
		return
	}

	// Extract XenblocksMiner
	srbMinerPath, err := extractSrbMiner(logView, logMessage, downloadedPath)
	if err != nil {
		return
	}

	logMessage(logView, fmt.Sprintf("srbMiner installed successfully at: %s", srbMinerPath))
}

func isSrbMinerInstalled(logView *tview.TextView, logMessage utils.LogMessageFunc) bool {
	var executableName string
	if runtime.GOOS == "windows" {
		executableName = "\\SRBMiner-Multi-2-6-6\\SRBMiner-MULTI.exe"
	} else {
		executableName = "/SRBMiner-Multi-2-6-6/SRBMiner-MULTI"
	}

	executablePath := filepath.Join(utils.GetExecutablePath(), SRB_MINER_DIR, executableName)
	if _, err := os.Stat(executablePath); err == nil {
		logMessage(logView, "srbMiner is already installed. No need to install again.")
		return true
	}
	return false
}

func downloadSrbMiner(logView *tview.TextView, logMessage utils.LogMessageFunc) (string, error) {
	var url, fileName string

	if runtime.GOOS == "windows" {
		url = "https://github.com/doktor83/SRBMiner-Multi/releases/download/2.6.6/SRBMiner-Multi-2-6-6-win64.zip"
		fileName = "SRBMiner-Multi-2-6-6-win64.zip"
	} else {
		url = "https://github.com/doktor83/SRBMiner-Multi/releases/download/2.6.6/SRBMiner-Multi-2-6-6-Linux.tar.gz"
		fileName = "SRBMiner-Multi-2-6-6-Linux.tar.gz"
	}

	// Construct the full file path
	filePath := filepath.Join(utils.GetExecutablePath(), SRB_MINER_DIR, fileName)

	// Download the file
	logMessage(logView, fmt.Sprintf("Downloading file from %s", url))
	resp, err := http.Get(url)
	if err != nil {
		logMessage(logView, fmt.Sprintf("Error downloading file: %v", err))
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Create the file
	logMessage(logView, fmt.Sprintf("Creating file: %s", filePath))
	out, err := os.Create(filePath)
	if err != nil {
		logMessage(logView, fmt.Sprintf("Error creating file: %v", err))
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Write the body to file
	logMessage(logView, "Saving downloaded content to file")
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logMessage(logView, fmt.Sprintf("Error saving file: %v", err))
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	logMessage(logView, fmt.Sprintf("File downloaded successfully to: %s", filePath))
	return filePath, nil
}

func extractSrbMiner(logView *tview.TextView, logMessage utils.LogMessageFunc, downloadedPath string) (string, error) {
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
		executablePath = filepath.Join(dir, "/SRBMiner-Multi-2-6-6/SRBMiner-MULTI")
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
		executablePath = filepath.Join(dir, "\\SRBMiner-Multi-2-6-6\\SRBMiner-MULTI.exe")
		// For Windows, no need to change permissions
		logMessage(logView, "File permissions update not required on Windows")
	}

	// Return the path to the executable
	return executablePath, nil
}
