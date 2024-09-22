package main

import (
	"xoon/ui"
	"xoon/utils"

	"github.com/rivo/tview"
)

var rootFlex *tview.Flex
var loginForm *tview.Form
var mainFlex *tview.Flex

func main() {
	utils.XoosInit()

	app := tview.NewApplication()
	rootFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	// Check for existing wallet
	publicKey := utils.CheckExistingWallet()

	if publicKey != "" {
		// Wallet exists, show login screen
		showLoginForm(app, publicKey)
	} else {
		// No wallet exists, show main interface directly
		showMainInterface(app)
	}

	ui.SetupInputCapture(app, func() {
		// Clean up all UI elements
		rootFlex.Clear()
		loginForm = nil
		mainFlex = nil
	})

	app.SetRoot(rootFlex, true).EnableMouse(true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func showLoginForm(app *tview.Application, publicKey string) {
	var passwordFieldIndex int

	loginForm = tview.NewForm().
		AddTextView("Instructions", "Please input password for existing encrypted wallet to unlock unmineable solXEN miner", 0, 2, false, false)

	passwordFieldIndex = loginForm.GetFormItemCount()
	loginForm.AddPasswordField("Password:", "", 32, '*', nil)

	loginForm.AddButton("Login", func() {
		password := loginForm.GetFormItem(passwordFieldIndex).(*tview.InputField).GetText()
		if utils.VerifyPassword(publicKey, password) {
			showMainInterface(app)
		} else {
			// Show error message
			//TODO
			// app.QueueUpdateDraw(func() {
			// 	// Check if error message already exists
			// 	errorIndex := loginForm.GetFormItemCount() - 1
			// 	if loginForm.GetFormItemCount() > passwordFieldIndex+2 {
			// 		loginForm.RemoveFormItem(errorIndex)
			// 	}
			// 	loginForm.AddTextView("Error:", "Invalid password", 0, 1, false, false)
			// })
		}
	}).
		AddButton("Exit", func() {
			app.Stop()
		})

	loginForm.SetBorder(true).SetTitle("Login to Wallet")
	rootFlex.Clear()
	rootFlex.AddItem(loginForm, 0, 1, true)
}

func showMainInterface(app *tview.Application) {
	mainMenu := ui.CreateMainMenu()
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	walletUI := ui.CreateWalletUI(app)
	solXENCPUUI := ui.CreateSolXENCPUUI(app)
	solXENGPUUI := ui.CreateSolXENGPUUI(app)

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

	mainFlex = tview.NewFlex().
		AddItem(mainMenu, 0, 1, true).
		AddItem(rightFlex, 0, 3, false)

	rootFlex.Clear()
	rootFlex.AddItem(mainFlex, 0, 1, true)
}
