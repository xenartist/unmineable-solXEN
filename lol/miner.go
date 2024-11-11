package xenblocks

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

var isMining bool = false

// Nvidia GPU algorithms
var GPUAlgorithms = []string{
	"FishHash (GPU>6GB)",
	"BeamHash (GPU>6GB)",
	"Ethash (GPU>6GB)",
	"Blake3 (GPU>4GB)",
	"Etchash (GPU>4GB)",
	"KarlsenHash (GPU>3GB)",
}

// Mining ports
var GPUMiningPorts = []string{"4444", "443", "3333", "13333", "80"}

func StartMining(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc,
	publicKey string, selectedAlgorithm string, selectedPort string, workerName string) {

	isMining = true

	go func() {
		var lolMinerPath = filepath.Join(utils.GetExecutablePath(), LOL_MINER_DIR)

		var err = os.Chdir(lolMinerPath)
		if err != nil {
			logMessage(logView, "Error changing to lolMiner directory: "+err.Error())
			return
		}

		// Set algorithm and host based on selectedAlgorithm
		var algorithm, host string
		switch selectedAlgorithm {
		case "KarlsenHash (GPU>3GB)":
			algorithm = "KARLSENV2"
			host = "karlsenhash.unmineable.com"
		case "Blake3 (GPU>4GB)":
			algorithm = "ALEPH"
			host = "blake3.unmineable.com"
		case "Etchash (GPU>4GB)":
			algorithm = "ETCHASH"
			host = "etchash.unmineable.com"
		case "FishHash (GPU>6GB)":
			algorithm = "FISHHASH"
			host = "fishhash.unmineable.com"
		case "BeamHash (GPU>6GB)":
			algorithm = "BEAM-III"
			host = "beamhash.unmineable.com"
		case "Ethash (GPU>6GB)":
			algorithm = "ETHASH"
			host = "ethash.unmineable.com"
		default:
			algorithm = "FISHHASH"
			host = "fishhash.unmineable.com"
		}

		// Construct the mining address
		miningAddress := host + ":" + selectedPort
		if selectedPort == "443" || selectedPort == "4444" {
			miningAddress = "stratum+ssl://" + miningAddress
		}

		// Construct the arguments slice
		args := []string{
			"--algo", algorithm,
			"--pool", miningAddress,
			"--user", fmt.Sprintf("SOL:%s.%s#plxp-imd8", publicKey, workerName),
		}

		var executableName string
		if runtime.GOOS == "windows" {
			executableName = ".\\1.91\\lolMiner.exe"
		} else {
			executableName = "./1.91/lolMiner"
		}

		cmd := exec.Command(executableName, args...)

		// Create pipes for both stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logMessage(logView, "Error creating StdoutPipe: "+err.Error())
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			logMessage(logView, "Error creating StderrPipe: "+err.Error())
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			logMessage(logView, "Error starting miner: "+err.Error())
			return
		}

		var (
			lastUpdateTime time.Time
			mutex          sync.Mutex
		)

		logMessage(logView, "Debug: Start Mining...Initiating takes a while...")

		// Function to read from a pipe and send to UI
		readPipe := func(pipe io.Reader) {
			reader := bufio.NewReader(pipe)

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					app.QueueUpdateDraw(func() {
						logMessage(logView, fmt.Sprintf("Error reading pipe: %v", err))
					})
					break
				}

				line = strings.TrimSpace(line)
				if line != "" {
					if strings.Contains(line, "Mining:") {
						mutex.Lock()
						now := time.Now()
						if now.Sub(lastUpdateTime) >= 10*time.Second {
							lastUpdateTime = now
							app.QueueUpdateDraw(func() {
								logMessage(logView, line)
							})
						}
						mutex.Unlock()
					} else if strings.Contains(line, "Ecosystem") {
						// skip
					} else {
						app.QueueUpdateDraw(func() {
							logMessage(logView, line)
						})
					}
				}

				// Force a UI update after each line
				app.QueueUpdateDraw(func() {})
			}
		}

		// Start goroutines to read from stdout and stderr
		go readPipe(stdout)
		go readPipe(stderr)

		// Wait for the command to finish
		if err := cmd.Wait(); err != nil {
			app.QueueUpdateDraw(func() {
				logMessage(logView, "Miner exited with error: "+err.Error())
			})
		} else {
			app.QueueUpdateDraw(func() {
				logMessage(logView, "Mining completed successfully")
			})
		}
	}()
}

func StopMining(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc) {
	KillMiningProcess()
	logMessage(logView, "Mining stopped")
	isMining = false

	// Change directory to parent
	if err := os.Chdir(".."); err != nil {
		logMessage(logView, "Error changing to parent directory: "+err.Error())
	}
}

// KillMiningProcess stops all running xenblocksMiner processes
func KillMiningProcess() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/F", "/IM", "lolMiner*")
	} else {
		cmd = exec.Command("pkill", "-f", "lolMiner")
	}
	_ = cmd.Run()
}

// IsMining returns the current mining status
func IsMining() bool {
	return isMining
}
