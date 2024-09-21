package ui

import (
	"xoon/utils"
	xenblocks "xoon/xmrig"

	"github.com/rivo/tview"
)

func CreateSolXENGPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("solXEN Miner (GPU)", app)

	// Ensure xenblocksMiner directory and config.txt exist

	if err := xenblocks.CreateXmrigMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating xenblocksMiner directory: "+err.Error())
	}

	// Create form
	publicKeyInput := tview.NewInputField().
		SetLabel("Solana Public Key").
		SetPlaceholder("Input public address").
		SetFieldWidth(46)

	form := tview.NewForm().
		AddFormItem(publicKeyInput)

	form.AddButton("Install Miner", func() { xenblocks.InstallXmrig(app, moduleUI.LogView, utils.LogMessage) }).
		AddButton("Start Mining", func() {
			if !xenblocks.IsMining() {
				publicKey := publicKeyInput.GetText()
				xenblocks.StartMining(app, moduleUI.LogView, utils.LogMessage, publicKey)
			}
		}).
		AddButton("Stop Mining", func() {
			if xenblocks.IsMining() {
				xenblocks.StopMining(app, moduleUI.LogView, utils.LogMessage)
			}
		})

	contentFlex := tview.NewFlex().AddItem(form, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENGPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("solXEN Config")
	return configFlex
}
