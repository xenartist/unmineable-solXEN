package ui

import (
	"fmt"
	"strconv"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

var tokenharvestForm *tview.Form

func CreateTokenHarvestUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("Token Harvest", app)

	// Create Auto Harvest form
	autoHarvestForm := tview.NewForm().
		SetBorder(true).
		SetTitle("Auto Harvest").
		SetTitleAlign(tview.AlignLeft)

	// Create Manual Harvest form
	var manualHarvestForm *tview.Form
	var solXenField *tview.InputField
	manualHarvestForm = tview.NewForm()

	manualHarvestForm.SetBorder(true).
		SetTitle("Manual Harvest").
		SetTitleAlign(tview.AlignLeft)

	// 1. Dropdown for SOL and input field
	solAmount := ""
	manualHarvestForm.AddInputField("SOL Amount", "", 20, nil, func(text string) {
		solAmount = text
		// Calculate solXEN amount when SOL amount changes
		go func() {
			if solAmount != "" {
				solAmountFloat, err := strconv.ParseFloat(solAmount, 64)
				if err != nil {
					utils.LogMessage(moduleUI.LogView, "Error parsing SOL amount: "+err.Error())
					return
				}
				solXenAmount, err := utils.GetTokenExchangeAmount(solAmountFloat, "solXEN")
				if err != nil {
					utils.LogMessage(moduleUI.LogView, "Error calculating solXEN amount: "+err.Error())
				} else {
					app.QueueUpdateDraw(func() {
						solXenField.SetText(fmt.Sprintf("%.6f", solXenAmount))
					})
				}
			} else {
				app.QueueUpdateDraw(func() {
					solXenField.SetText("")
				})
			}
		}()
	})

	// 2. Dropdown for solXEN and input field (readonly)
	solXenField = tview.NewInputField()
	solXenField.SetLabel("solXEN Amount").
		SetText("").
		SetFieldWidth(20).
		SetDisabled(true)
	manualHarvestForm.AddFormItem(solXenField)

	// 3. Swap button
	manualHarvestForm.AddButton("Harvest", func() {

		result, err := utils.ExchangeSolForToken(solAmount, "solXEN")
		if err != nil {
			// Handle error
			utils.LogMessage(moduleUI.LogView, "Error: "+err.Error())
		} else {
			solXenField.SetText(result)
			utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Swapped %s SOL for solXEN: %s successfully", solAmount, result))

			// Update wallet info after 60 seconds
			go func() {
				time.Sleep(60 * time.Second)
				UpdateWalletInfo(app, walletInfoView)
			}()
		}

	})

	// Create a flex container for the two forms
	formsFlex := tview.NewFlex().
		AddItem(autoHarvestForm, 0, 1, true).
		AddItem(manualHarvestForm, 0, 1, false)

	// Add the forms flex to the content flex
	contentFlex := tview.NewFlex().AddItem(formsFlex, 0, 1, true)

	moduleUI.ConfigFlex.AddItem(contentFlex, 0, 1, true)

	return moduleUI
}

func CreateTokenHarvestConfigFlex(app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle("Token Harvest Config")
	return configFlex
}
