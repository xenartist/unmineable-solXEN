package ui

import (
	"xoon/utils"

	"github.com/atotto/clipboard"
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
		AddButton("Copy Public Key", func() {
			err := clipboard.WriteAll(utils.GLOBAL_PUBLIC_KEY)
			if err != nil {
				utils.LogMessage(moduleUI.LogView, "Error copying public key to clipboard: "+err.Error())
			} else {
				utils.LogMessage(moduleUI.LogView, "Public key copied to clipboard successfully")
			}
		}).
		AddButton("Copy Private Key", func() {
			err := clipboard.WriteAll(utils.GLOBAL_PRIVATE_KEY)
			if err != nil {
				utils.LogMessage(moduleUI.LogView, "Error copying private key to clipboard: "+err.Error())
			} else {
				utils.LogMessage(moduleUI.LogView, "Private key copied to clipboard successfully")
			}
		})
	manageWalletForm.SetBorder(true).SetTitle("Manage Wallet")

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
