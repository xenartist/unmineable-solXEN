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
	LOL_MINER_DIR = "lolMiner"
)

func CreateLolMinerDir(logView *tview.TextView, logMessage utils.LogMessageFunc) error {
	lolMinerPath := filepath.Join(utils.GetExecutablePath(), LOL_MINER_DIR)
	if _, err := os.Stat(lolMinerPath); os.IsNotExist(err) {
		err = os.Mkdir(lolMinerPath, 0755)
		if err != nil {
			logMessage(logView, fmt.Sprintf("Error creating lolMiner directory: %v", err))
			return err
		}
		logMessage(logView, "lolMiner directory created successfully")
	}

	return nil
}

func InstallLolMiner(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc) {

	// Log the start of the installation process
	if isLolMinerInstalled(logView, logMessage) {
		return
	}

	logMessage(logView, "Starting lolMiner installation...")

	downloadedPath, err := downloadLolMiner(logView, logMessage)
	if err != nil {
		return
	}

	// Extract XenblocksMiner
	lolMinerPath, err := extractLolMiner(logView, logMessage, downloadedPath)
	if err != nil {
		return
	}

	logMessage(logView, fmt.Sprintf("lolMiner installed successfully at: %s", lolMinerPath))
}

func isLolMinerInstalled(logView *tview.TextView, logMessage utils.LogMessageFunc) bool {
	var executableName string
	if runtime.GOOS == "windows" {
		executableName = "\\1.91\\lolMiner.exe"
	} else {
		executableName = "/1.91/lolMiner"
	}

	executablePath := filepath.Join(utils.GetExecutablePath(), LOL_MINER_DIR, executableName)
	if _, err := os.Stat(executablePath); err == nil {
		logMessage(logView, "lolMiner is already installed. No need to install again.")
		return true
	}
	return false
}

func downloadLolMiner(logView *tview.TextView, logMessage utils.LogMessageFunc) (string, error) {
	var url, fileName string

	if runtime.GOOS == "windows" {
		url = "https://github.com/Lolliedieb/lolMiner-releases/releases/download/1.91/lolMiner_v1.91_Win64.zip"
		fileName = "lolMiner_v1.91_Win64.zip"
	} else {
		url = "https://github.com/Lolliedieb/lolMiner-releases/releases/download/1.91/lolMiner_v1.91_Lin64.tar.gz"
		fileName = "lolMiner_v1.91_Lin64.tar.gz"
	}

	// Construct the full file path
	filePath := filepath.Join(utils.GetExecutablePath(), LOL_MINER_DIR, fileName)

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

func extractLolMiner(logView *tview.TextView, logMessage utils.LogMessageFunc, downloadedPath string) (string, error) {
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
		executablePath = filepath.Join(dir, "/1.91/lolMiner")
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
		executablePath = filepath.Join(dir, "\\1.91\\lolMiner.exe")
		// For Windows, no need to change permissions
		logMessage(logView, "File permissions update not required on Windows")
	}

	// Return the path to the executable
	return executablePath, nil
}
