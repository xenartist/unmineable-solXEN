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

// CPU algorithms
var CPUAlgorithms = []string{"GhostRider", "RandomX"}

// Mining ports
var CPUMiningPorts = []string{"443", "3333", "13333", "80"}

func StartMining(app *tview.Application, logView *tview.TextView, logMessage utils.LogMessageFunc,
	publicKey string, selectedThreads string, selectedAlgorithm string, selectedPort string, workerName string) {

	isMining = true

	go func() {
		var xenblocksMinerPath = filepath.Join(utils.GetExecutablePath(), XMRIG_MINER_DIR)

		var err = os.Chdir(xenblocksMinerPath)
		if err != nil {
			logMessage(logView, "Error changing to xenblocksMiner directory: "+err.Error())
			return
		}

		// Set algorithm and host based on selectedAlgorithm
		var algorithm, host string
		switch selectedAlgorithm {
		case "GhostRider":
			algorithm = "gr"
			host = "ghostrider.unmineable.com"
		case "RandomX":
			algorithm = "rx"
			host = "rx.unmineable.com"
		default:
			// Default to GhostRider if unknown algorithm is provided
			algorithm = "gr"
			host = "ghostrider.unmineable.com"
		}

		// Construct the mining address
		miningAddress := host + ":" + selectedPort
		if selectedPort == "443" {
			miningAddress = "stratum+ssl://" + miningAddress
		}

		// Construct the arguments slice
		args := []string{
			"-a", algorithm,
			"-t", selectedThreads,
			"-o", miningAddress,
			"-u", fmt.Sprintf("SOL:%s.%s#plxp-imd8", publicKey, workerName),
			"-p", "x",
		}

		var executableName string
		if runtime.GOOS == "windows" {
			executableName = ".\\xmrig-6.22.0\\xmrig.exe"
		} else {
			executableName = "./xmrig-6.22.0/xmrig"
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

		logMessage(logView, "Debug: Start Mining function called")

		// Function to read from a pipe and send to UI
		readPipe := func(pipe io.Reader) {
			reader := bufio.NewReader(pipe)
			buffer := make([]byte, 1024)

			logMessage(logView, "Debug: Starting to read pipe")

			for {
				n, err := reader.Read(buffer)
				if err != nil {
					if err == io.EOF {
						logMessage(logView, "Debug: EOF reached, waiting...")
						time.Sleep(100 * time.Millisecond)
						continue
					}
					logMessage(logView, fmt.Sprintf("Error reading pipe: %v", err))
					break
				}

				if n > 0 {
					output := string(buffer[:n])
					//logMessage(logView, fmt.Sprintf("Debug: Read %d bytes", n))

					lines := strings.Split(output, "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" {
							if strings.Contains(line, "Mining:") {
								mutex.Lock()
								now := time.Now()
								if now.Sub(lastUpdateTime) >= 10*time.Second {
									lastUpdateTime = now
									logMessage(logView, line)
								}
								mutex.Unlock()
							} else if strings.Contains(line, "Ecosystem") {
								//skip
							} else {
								logMessage(logView, line)
							}
						}
					}
				} else {
					logMessage(logView, "Debug: No data read")
				}
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
		cmd = exec.Command("taskkill", "/F", "/IM", "xmrig*")
	} else {
		cmd = exec.Command("pkill", "-f", "xmrig")
	}
	_ = cmd.Run()
}

// IsMining returns the current mining status
func IsMining() bool {
	return isMining
}
