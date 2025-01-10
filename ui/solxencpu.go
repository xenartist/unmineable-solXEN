package ui

import (
	"xoon/utils"
	xenblocks "xoon/xmrig"

	"runtime"
	"strconv"

	"github.com/rivo/tview"
)

var solxencpuForm *tview.Form = tview.NewForm()

// Add this helper function to generate thread options
func generateThreadOptions() []string {
	maxThreads := runtime.NumCPU()
	options := []string{"1"} // Start with 1 thread

	for threads := 2; threads <= maxThreads; threads *= 2 {
		options = append(options, strconv.Itoa(threads))
	}
	return options
}

func CreateSolXENCPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI(SOLXEN_CPU_MINER_STRING, app)

	// Ensure xenblocksMiner directory and config.txt exist
	if err := xenblocks.CreateXmrigMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating xenblocksMiner directory: "+err.Error())
	}

	// Determine the public key display text
	publicKeyDisplay := ""
	if utils.GetGlobalPublicKey() != "" {
		publicKeyDisplay = utils.GetGlobalPublicKey()[:8] + "********"
	}

	var selectedAlgorithm, selectedPort, workerName string
	var selectedThreads string = "1" // Default value

	solxencpuForm.AddTextView("Public Key", publicKeyDisplay, 0, 1, false, true).
		AddDropDown("CPU Threads", generateThreadOptions(), 0, func(option string, index int) {
			selectedThreads = option
		}).
		AddDropDown("Mining Algorithm", xenblocks.CPUAlgorithms, 0, func(option string, index int) {
			selectedAlgorithm = option
		}).
		AddDropDown("Port", xenblocks.CPUMiningPorts, 0, func(option string, index int) {
			selectedPort = option
		}).
		AddInputField("Worker Name", "xoon", 10, nil, func(text string) {
			workerName = text
		}).
		AddButton("Install Miner", func() { xenblocks.InstallXmrig(app, moduleUI.LogView, utils.LogMessage) }).
		AddButton("Start Mining", func() {
			if !xenblocks.IsMining() {
				publicKey := utils.GetGlobalPublicKey()
				xenblocks.StartMining(app, moduleUI.LogView, utils.LogMessage,
					publicKey, selectedThreads, selectedAlgorithm, selectedPort, workerName)
			}
		}).
		AddButton("Stop Mining", func() {
			if xenblocks.IsMining() {
				xenblocks.StopMining(app, moduleUI.LogView, utils.LogMessage)
			}
		})

	contentFlex := tview.NewFlex().AddItem(solxencpuForm, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENCPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle(SOLXEN_CPU_MINER_STRING + " Config")
	return configFlex
}

func UpdateCPUMinerPublicKeyTextView() {
	if solxencpuForm == nil {
		return
	}

	if utils.GetGlobalPublicKey() == "" {
		solxencpuForm.GetFormItem(0).(*tview.TextView).SetText("")
	} else {
		solxencpuForm.GetFormItem(0).(*tview.TextView).SetText(utils.GetGlobalPublicKey()[:8] + "********")
	}
}
