package ui

import (
	"xoon/utils"
	xenblocks "xoon/xmrig"

	"github.com/rivo/tview"
)

func CreateSolXENCPUUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("solXEN Miner (CPU)", app)

	// Ensure xenblocksMiner directory and config.txt exist
	if err := xenblocks.CreateXmrigMinerDir(moduleUI.LogView, utils.LogMessage); err != nil {
		utils.LogMessage(moduleUI.LogView, "Error creating xenblocksMiner directory: "+err.Error())
	}

	// Create form
	form := tview.NewForm()

	form.AddTextView("Public Key", utils.GLOBAL_PUBLIC_KEY, 0, 1, false, true).
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

	contentFlex := tview.NewFlex().AddItem(form, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateSolXENCPUConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("solXEN Config")
	return configFlex
}
