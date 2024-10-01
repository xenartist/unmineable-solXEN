package ui

import (
	"fmt"
	"strconv"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

//var tokenharvestForm *tview.Form

func CreateTokenHarvestUI(app *tview.Application) ModuleUI {
	var moduleUI = CreateModuleUI("Token Harvest", app)

	autoHarvestForm := createAutoHarvestForm(app, &moduleUI, walletInfoView)
	manualHarvestForm := createManualHarvestForm(app, &moduleUI, walletInfoView)

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

func createAutoHarvestForm(app *tview.Application, moduleUI *ModuleUI, walletInfoView *tview.TextView) *tview.Form {
	// Create Auto Harvest form
	autoHarvestForm := tview.NewForm()
	autoHarvestForm.SetBorder(true).
		SetTitle("Auto Harvest").
		SetTitleAlign(tview.AlignLeft)

	// Read configuration from file
	config, err := utils.ReadSolXENConfigFile()
	if err != nil {
		utils.LogToFile("Failed to read config file: " + err.Error())
		// Use default values if config file can't be read
		config = utils.SolXENConfig{
			AutoHarvestActive: true,
			SOLPerHarvest:     0.01,
			TokenToHarvest:    "solXEN",
			HarvestInterval:   "1h",
		}
	}

	// 1. Checkbox for auto harvest activation
	autoHarvestForm.AddCheckbox("Auto Harvest (ON/OFF)", config.AutoHarvestActive, func(checked bool) {
		config.AutoHarvestActive = checked
	})

	// 2. Input field for SOL amount per harvest
	solAmountPerHarvest := strconv.FormatFloat(config.SOLPerHarvest, 'f', -1, 64)
	autoHarvestForm.AddInputField("SOL per Harvest", solAmountPerHarvest, 10, nil, func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			config.SOLPerHarvest = val
		}
	})

	// 3. Dropdown for token selection
	tokenOptions := []string{"solXEN", "xencat", "ORE"}
	tokenIndex := 0
	for i, token := range tokenOptions {
		if token == config.TokenToHarvest {
			tokenIndex = i
			break
		}
	}
	autoHarvestForm.AddDropDown("Token to Harvest", tokenOptions, tokenIndex, func(option string, index int) {
		config.TokenToHarvest = option
	})

	// 4. Dropdown for harvest interval
	intervalOptions := []string{"10m", "1h", "1d"}
	intervalIndex := 1 // default to 1h
	for i, interval := range intervalOptions {
		if interval == config.HarvestInterval {
			intervalIndex = i
			break
		}
	}
	autoHarvestForm.AddDropDown("Harvest Interval", intervalOptions, intervalIndex, func(option string, index int) {
		config.HarvestInterval = option
	})

	// Add a channel to trigger config reload
	reloadConfigChan := make(chan struct{})

	// 5. Save Config button
	autoHarvestForm.AddButton("Save Config & Restart", func() {
		err := utils.WriteSolXENConfigFile(config)
		if err != nil {
			utils.LogMessage(moduleUI.LogView, "Failed to save config: "+err.Error())
		} else {
			utils.LogMessage(moduleUI.LogView, "Auto Harvest configuration saved")
			// Trigger config reload
			reloadConfigChan <- struct{}{}
		}
	})

	go func() {
		var ticker *time.Ticker
		defer func() {
			if ticker != nil {
				ticker.Stop()
			}
		}()

		for {
			config, err := utils.ReadSolXENConfigFile()
			if err != nil {
				utils.LogMessage(moduleUI.LogView, "Failed to read config: "+err.Error())
				time.Sleep(1 * time.Hour) // Default sleep if config read fails
				continue
			}

			// Parse the harvest interval
			var interval time.Duration
			switch config.HarvestInterval {
			case "10m":
				interval = 10 * time.Minute
				utils.LogMessage(moduleUI.LogView, "Time until next harvest: 10m")
			case "1h":
				interval = 1 * time.Hour
				utils.LogMessage(moduleUI.LogView, "Time until next harvest: 1h")
			case "1d":
				interval = 24 * time.Hour
				utils.LogMessage(moduleUI.LogView, "Time until next harvest: 1d")
			default:
				utils.LogMessage(moduleUI.LogView, "Invalid harvest interval: "+config.HarvestInterval+". Using default 1 hour.")
				interval = 1 * time.Hour
			}

			// Create or update the ticker
			if ticker != nil {
				ticker.Stop()
			}
			ticker = time.NewTicker(interval)

			// Start time for countdown
			startTime := time.Now()

			// Countdown ticker
			var countdownInterval time.Duration
			switch config.HarvestInterval {
			case "10m":
				countdownInterval = 2 * time.Minute
			case "1h":
				countdownInterval = 10 * time.Minute
			case "1d":
				countdownInterval = 2 * time.Hour
			default:
				countdownInterval = 10 * time.Minute
			}

			countdownTicker := time.NewTicker(countdownInterval)
			defer countdownTicker.Stop()

		counterdownLoop:
			for {
				select {
				case <-ticker.C:
					if config.AutoHarvestActive {
						// Execute token swap based on configuration
						result, err := utils.ExchangeSolForToken(strconv.FormatFloat(config.SOLPerHarvest, 'f', -1, 64), config.TokenToHarvest)
						if err != nil {
							utils.LogMessage(moduleUI.LogView, "Error: "+err.Error())
						} else {
							utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Swapped %v SOL for %s: %s successfully", config.SOLPerHarvest, config.TokenToHarvest, result))

							// Update wallet info after 60 seconds
							go func() {
								time.Sleep(60 * time.Second)
								UpdateWalletInfo(app, walletInfoView)
							}()
						}
					}
					break counterdownLoop

				case <-countdownTicker.C:
					remainingTime := interval - time.Since(startTime)
					if remainingTime > 0 {
						utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Time until next harvest: %s", remainingTime.Round(time.Minute)))
					}

				case <-reloadConfigChan:
					// Break the loop to reload config immediately
					utils.LogMessage(moduleUI.LogView, "Reloading configuration...")
					break counterdownLoop
				}

				// Check if it's time to break the loop and re-read config
				if time.Since(startTime) >= interval {
					break
				}
			}
		}
	}()

	return autoHarvestForm
}

func createManualHarvestForm(app *tview.Application, moduleUI *ModuleUI, walletInfoView *tview.TextView) *tview.Form {
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
		utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Swapping %s SOL for solXEN", solAmount))

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

	return manualHarvestForm
}
