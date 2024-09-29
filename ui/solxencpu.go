package ui

import (
	"fmt"
	"xoon/utils"
	xenblocks "xoon/xmrig"

	"github.com/rivo/tview"
)

var solxencpuForm *tview.Form

func CreateSolXENCPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("solXEN Miner (CPU)", app)

	// Ensure xenblocksMiner directory and config.txt exist
	if err := xenblocks.CreateXmrigMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating xenblocksMiner directory: "+err.Error())
	}

	// Create form
	solxencpuForm = tview.NewForm()

	// Determine the public key display text
	publicKeyDisplay := ""
	if utils.GetGlobalPublicKey() != "" {
		publicKeyDisplay = utils.GetGlobalPublicKey()[:8] + "********"
	}

	var selectedAlgorithm, selectedPort, workerName string

	solxencpuForm.AddTextView("Public Key", publicKeyDisplay, 0, 1, false, true).
		AddDropDown("Mining Algorithm", xenblocks.CPUAlgorithms, 0, func(option string, index int) {
			selectedAlgorithm = option
		}).
		AddDropDown("Port", xenblocks.MiningPorts, 0, func(option string, index int) {
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
					publicKey, selectedAlgorithm, selectedPort, workerName)
			}
		}).
		AddButton("Stop Mining", func() {
			if xenblocks.IsMining() {
				xenblocks.StopMining(app, moduleUI.LogView, utils.LogMessage)
			}
		}).
		AddButton("Test Swap", func() {
			go func() {
				amount, err := utils.ExchangeSolForToken("0.001", "solXEN")
				if err != nil {
					utils.LogMessage(moduleUI.LogView, "Swap failed: "+err.Error())
				} else {
					utils.LogMessage(moduleUI.LogView, "Swapped SOL for solXEN "+fmt.Sprintf("%f", amount))
				}
			}()
		})

	contentFlex := tview.NewFlex().AddItem(solxencpuForm, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENCPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("solXEN Config")
	return configFlex
}

func UpdateCPUMinerPublicKeyTextView() {
	if utils.GetGlobalPublicKey() == "" {
		solxencpuForm.GetFormItem(0).(*tview.TextView).SetText("")
	} else {
		solxencpuForm.GetFormItem(0).(*tview.TextView).SetText(utils.GetGlobalPublicKey()[:8] + "********")
	}
}
