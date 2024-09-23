package ui

import (
	"xoon/utils"

	"github.com/rivo/tview"
)

func CreateWalletUI(app *tview.Application) ModuleUI {
	moduleUI := CreateModuleUI("Solana Wallet", app)

	var createWalletForm *tview.Form
	var manageWalletForm *tview.Form

	// Create Wallet form
	createWalletForm = tview.NewForm().
		AddPasswordField("Password (min 8 characters):", "", 32, '*', nil).
		AddPasswordField("Confirm Password:", "", 32, '*', nil).
		AddButton("Create Wallet", func() {
			password := createWalletForm.GetFormItem(0).(*tview.InputField).GetText()
			confirm := createWalletForm.GetFormItem(1).(*tview.InputField).GetText()

			if len(password) < 8 {
				moduleUI.LogView.SetText("Password must be at least 8 characters long\n")
				return
			}
			if password != confirm {
				moduleUI.LogView.SetText("Passwords do not match\n")
				return
			}

			utils.CreateNewWallet(app, moduleUI.LogView, utils.LogMessage, password)

			// After successful wallet creation, update the TextView
			go func() {
				// Hide createWalletForm
				createWalletForm.Clear(true)

				newPublicKey := utils.GLOBAL_PUBLIC_KEY
				app.QueueUpdateDraw(func() {
					publicKeyTextView := manageWalletForm.GetFormItemByLabel("Public Key").(*tview.TextView)
					publicKeyTextView.SetText(newPublicKey)
				})
			}()
		})
	createWalletForm.SetBorder(true).SetTitle("Create Wallet")

	// Manage Wallet form
	manageWalletForm = tview.NewForm()
	manageWalletForm.
		AddTextView("Public Key", utils.GLOBAL_PUBLIC_KEY, 0, 1, false, true).
		AddPasswordField("Input password to export wallet:", "", 32, '*', nil).
		AddButton("Export Wallet", func() {
			password := manageWalletForm.GetFormItem(1).(*tview.InputField).GetText()
			if password == "" {
				utils.LogMessage(moduleUI.LogView, "Please enter your password to export the wallet")
				return
			} else if password != utils.GLOBAL_PASSWORD {
				utils.LogMessage(moduleUI.LogView, "Incorrect password")
			} else {
				err := utils.ExportWallet()
				if err != nil {
					utils.LogMessage(moduleUI.LogView, "Error exporting wallet: "+err.Error())
				} else {
					utils.LogMessage(moduleUI.LogView, "Wallet exported successfully under wallet folder")
					manageWalletForm.GetFormItem(1).(*tview.InputField).SetText("")
				}
			}
		})
	manageWalletForm.SetBorder(true).SetTitle("Manage Wallet")

	// Create a flex layout for vertical arrangement
	walletFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	if utils.GLOBAL_PUBLIC_KEY != "" {
		// If GLOBAL_PUBLIC_KEY is not empty, only add manageWalletForm
		walletFlex.AddItem(manageWalletForm, 0, 1, true)
	} else {
		// If GLOBAL_PUBLIC_KEY is empty, add both forms
		walletFlex.
			AddItem(createWalletForm, 0, 1, true).
			AddItem(manageWalletForm, 0, 1, false)
	}

	// Add the flex layout to moduleUI
	moduleUI.ConfigFlex.AddItem(walletFlex, 0, 1, true)

	return moduleUI
}

func CreateWalletConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("Solana Wallet")
	return configFlex
}
