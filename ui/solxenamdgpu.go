package ui

import (
	xenblocks "xoon/srb"
	"xoon/utils"

	"github.com/rivo/tview"
)

var solxenamdgpuForm *tview.Form = tview.NewForm()

func CreateSolXENAMDGPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI(SOLXEN_AMD_GPU_MINER_STRING, app)

	if err := xenblocks.CreateSrbMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating srbMiner directory: "+err.Error())
	}

	// Determine the public key display text
	publicKeyDisplay := ""
	if utils.GetGlobalPublicKey() != "" {
		publicKeyDisplay = utils.GetGlobalPublicKey()[:8] + "********"
	}

	var selectedAlgorithm, selectedPort, workerName string

	solxenamdgpuForm.AddTextView("Public Key", publicKeyDisplay, 0, 1, false, true).
		AddDropDown("Mining Algorithm", xenblocks.GPUAlgorithms, 0, func(option string, index int) {
			selectedAlgorithm = option
		}).
		AddDropDown("Port", xenblocks.GPUMiningPorts, 0, func(option string, index int) {
			selectedPort = option
		}).
		AddInputField("Worker Name", "xoon", 10, nil, func(text string) {
			workerName = text
		}).
		AddButton("Install Miner", func() { xenblocks.InstallSrbMiner(app, moduleUI.LogView, utils.LogMessage) }).
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
		})

	contentFlex := tview.NewFlex().AddItem(solxenamdgpuForm, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENAMDGPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle(SOLXEN_AMD_GPU_MINER_STRING + " Config")
	return configFlex
}

func UpdateAMDGPUMinerPublicKeyTextView() {
	if solxenamdgpuForm == nil {
		return
	}

	if utils.GetGlobalPublicKey() == "" {
		solxenamdgpuForm.GetFormItem(0).(*tview.TextView).SetText("")
	} else {
		solxenamdgpuForm.GetFormItem(0).(*tview.TextView).SetText(utils.GetGlobalPublicKey()[:8] + "********")
	}
}
