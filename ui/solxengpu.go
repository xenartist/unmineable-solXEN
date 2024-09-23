package ui

import (
	"xoon/utils"
	xenblocks "xoon/xmrig"

	"github.com/rivo/tview"
)

var solxengpuForm *tview.Form

func CreateSolXENGPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("solXEN Miner (GPU)", app)

	// Ensure xenblocksMiner directory and config.txt exist

	if err := xenblocks.CreateXmrigMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating xenblocksMiner directory: "+err.Error())
	}

	// Create form
	solxengpuForm := tview.NewForm()

	// Determine the public key display text
	publicKeyDisplay := ""
	if utils.GLOBAL_PUBLIC_KEY != "" {
		publicKeyDisplay = utils.GLOBAL_PUBLIC_KEY[:8] + "********"
	}

	solxengpuForm.AddTextView("Public Key", publicKeyDisplay, 0, 1, false, true).
		AddButton("Install Miner", func() { xenblocks.InstallXmrig(app, moduleUI.LogView, utils.LogMessage) }).
		AddButton("Start Mining", func() {
			if !xenblocks.IsMining() {
				publicKey := utils.GLOBAL_PUBLIC_KEY
				xenblocks.StartMining(app, moduleUI.LogView, utils.LogMessage, publicKey)
			}
		}).
		AddButton("Stop Mining", func() {
			if xenblocks.IsMining() {
				xenblocks.StopMining(app, moduleUI.LogView, utils.LogMessage)
			}
		})

	contentFlex := tview.NewFlex().AddItem(solxengpuForm, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENGPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("solXEN Config")
	return configFlex
}

func UpdateGPUMinerPublicKeyTextView() {
	if utils.GLOBAL_PUBLIC_KEY == "" {
		solxengpuForm.GetFormItem(0).(*tview.TextView).SetText("")
	} else {
		solxengpuForm.GetFormItem(0).(*tview.TextView).SetText(utils.GLOBAL_PUBLIC_KEY[:8] + "********")
	}
}
