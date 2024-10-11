package ui

import (
	xenblocks "xoon/lol"
	"xoon/utils"

	"github.com/rivo/tview"
)

var solxennvidiaForm *tview.Form = tview.NewForm()

func CreateSolXENNvidiaGPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI(SOLXEN_NVIDIA_GPU_MINER_STRING, app)

	if err := xenblocks.CreateLolMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating lolMiner directory: "+err.Error())
	}

	// Determine the public key display text
	publicKeyDisplay := ""
	if utils.GetGlobalPublicKey() != "" {
		publicKeyDisplay = utils.GetGlobalPublicKey()[:8] + "********"
	}

	var selectedAlgorithm, selectedPort, workerName string

	solxennvidiaForm.AddTextView("Public Key", publicKeyDisplay, 0, 1, false, true).
		AddDropDown("Mining Algorithm", xenblocks.GPUAlgorithms, 0, func(option string, index int) {
			selectedAlgorithm = option
		}).
		AddDropDown("Port", xenblocks.GPUMiningPorts, 0, func(option string, index int) {
			selectedPort = option
		}).
		AddInputField("Worker Name", "xoon", 10, nil, func(text string) {
			workerName = text
		}).
		AddButton("Install Miner", func() { xenblocks.InstallLolMiner(app, moduleUI.LogView, utils.LogMessage) }).
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

	contentFlex := tview.NewFlex().AddItem(solxennvidiaForm, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENNvidiaGPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle(SOLXEN_NVIDIA_GPU_MINER_STRING + " Config")
	return configFlex
}

func UpdateNvidiaGPUMinerPublicKeyTextView() {
	if solxennvidiaForm == nil {
		return
	}

	if utils.GetGlobalPublicKey() == "" {
		solxennvidiaForm.GetFormItem(0).(*tview.TextView).SetText("")
	} else {
		solxennvidiaForm.GetFormItem(0).(*tview.TextView).SetText(utils.GetGlobalPublicKey()[:8] + "********")
	}
}
