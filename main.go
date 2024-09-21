package main

import (
	"xoon/ui"
	"xoon/utils"

	"github.com/rivo/tview"
)

func main() {
	utils.XoosInit()

	app := tview.NewApplication()

	mainMenu := ui.CreateMainMenu()
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	walletUI := ui.CreateWalletUI(app)
	solXENCPUUI := ui.CreateSolXENCPUUI(app)
	solXENGPUUI := ui.CreateSolXENGPUUI(app)

	//click items in mainmenu to swith views

	switchView := ui.CreateSwitchViewFunc(rightFlex, mainMenu)

	modules := []ui.ModuleUI{
		{
			DashboardFlex: walletUI.DashboardFlex,
			ConfigFlex:    walletUI.ConfigFlex,
			LogView:       walletUI.LogView,
		},
		{
			DashboardFlex: solXENCPUUI.DashboardFlex,
			ConfigFlex:    solXENCPUUI.ConfigFlex,
			LogView:       solXENCPUUI.LogView,
		},
		{
			DashboardFlex: solXENGPUUI.DashboardFlex,
			ConfigFlex:    solXENGPUUI.ConfigFlex,
			LogView:       solXENGPUUI.LogView,
		},
	}

	ui.SetupMenuItemSelection(mainMenu, switchView, modules)

	mainFlex := tview.NewFlex().
		AddItem(mainMenu, 0, 1, true).
		AddItem(rightFlex, 0, 3, false)

	//input capture, eg. press 4 times q to quit
	ui.SetupInputCapture(app)

	if err := app.SetRoot(mainFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
