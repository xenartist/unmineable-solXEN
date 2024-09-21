package ui

import (
	"xoon/utils"

	"github.com/rivo/tview"
)

func CreateWalletUI(app *tview.Application) ModuleUI {
	moduleUI := CreateModuleUI("Solana Wallet", app)

	var createWalletForm *tview.Form

	// Create Wallet form
	createWalletForm = tview.NewForm().
		AddPasswordField("Password (min 8 characters):", "", 0, '*', nil).
		AddPasswordField("Confirm Password:", "", 0, '*', nil).
		AddButton("Create Wallet", func() {
			password := createWalletForm.GetFormItem(0).(*tview.InputField).GetText()
			confirm := createWalletForm.GetFormItem(1).(*tview.InputField).GetText()

			if len(password) < 8 {
				moduleUI.LogView.SetText("Password must be at least 8 characters long")
				return
			}
			if password != confirm {
				moduleUI.LogView.SetText("Passwords do not match")
				return
			}

			utils.CreateNewWallet(app, moduleUI.LogView, utils.LogMessage, password)
		})
	createWalletForm.SetBorder(true).SetTitle("Create Wallet")

	// Manage Wallet form
	manageWalletForm := tview.NewForm()
	manageWalletForm.
		AddButton("Copy Public Key", func() {
			// TODO: Implement copy to clipboard
		}).
		AddPasswordField("Password:", "", 0, '*', nil).
		AddButton("Copy Private Key", func() {
			// TODO: Implement decryption and copy to clipboard
		})
	manageWalletForm.SetBorder(true).SetTitle("Manage Wallet")

	// Update public key display
	updatePublicKeyDisplay := func() {
		// TODO: Read wallet file and update publicKeyText
	}
	updatePublicKeyDisplay()

	// Create a flex layout for vertical arrangement
	walletFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(createWalletForm, 0, 1, true).
		AddItem(manageWalletForm, 0, 1, false)

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
