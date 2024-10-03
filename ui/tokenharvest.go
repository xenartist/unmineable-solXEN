package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

var tokenOptions = []string{"solXEN", "xencat", "ORE"}

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
			// AutoHarvestActive: true,
			SOLPerHarvest:   0.01,
			TokenToHarvest:  "solXEN",
			HarvestInterval: "1h",
		}
	}

	// 1. Checkbox for auto harvest activation
	// autoHarvestForm.AddCheckbox("Auto Harvest (ON/OFF)", config.AutoHarvestActive, func(checked bool) {
	// 	config.AutoHarvestActive = checked
	// })

	// 2. Input field for SOL amount per harvest
	solAmountPerHarvest := strconv.FormatFloat(config.SOLPerHarvest, 'f', -1, 64)
	autoHarvestForm.AddInputField("SOL per Harvest", solAmountPerHarvest, 10, func(textToCheck string, lastChar rune) bool {
		// Only allow digits and one decimal point
		if (lastChar >= '0' && lastChar <= '9') || (lastChar == '.' && strings.Contains(textToCheck, ".")) {
			return true
		}
		return false
	}, func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			config.SOLPerHarvest = val
		}
	})

	// 3. Dropdown for token selection
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
	autoHarvestForm.AddButton("Save Config & Let's Fuckin Harvest!", func() {
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
					// Check wallet balance
					balances, err := utils.GetWalletTokenBalances(utils.GetGlobalPublicKey())
					if err != nil {
						utils.LogMessage(moduleUI.LogView, "Error getting wallet balances: "+err.Error())
						break counterdownLoop
					}

					// Get SOL balance
					solBalance, err := utils.GetSOLBalance(utils.GetGlobalPublicKey())
					if err != nil {
						utils.LogMessage(moduleUI.LogView, "Error getting SOL balance: "+err.Error())
						break counterdownLoop
					}

					// Get solXEN balance
					solXENBalance := 0.0
					for _, balance := range balances {
						if balance.Symbol == "solXEN" {
							solXENBalance = balance.Balance
							break
						}
					}

					if config.TokenToHarvest != "solXEN" {

						// Check if SOL balance is sufficient
						if solBalance < 0.000006 {
							utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Insufficient SOL balance %f. Minimum required: 0.000006 SOL", solBalance))
							break counterdownLoop
						}

						if solXENBalance < 42069 {
							// Calculate how many solXEN to buy
							solXENToBuy := 46920 - solXENBalance

							// Calculate SOL amount needed for solXEN
							solRequiredAmount, err := utils.GetSolExchangeAmount(fmt.Sprintf("%f", solXENToBuy), "solXEN")
							if err != nil {
								utils.LogMessage(moduleUI.LogView, "Error calculating SOL amount for solXEN: "+err.Error())
								break counterdownLoop
							}

							// Ensure minimum SOL amount
							solRequiredAmountFloat, _ := strconv.ParseFloat(solRequiredAmount, 64)
							if solBalance < solRequiredAmountFloat+0.000005 {
								utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Insufficient SOL balance %f. Minimum required: %.9f SOL", solBalance, solRequiredAmountFloat+0.000005))
								break counterdownLoop
							}

							// Ensure minimum SOL amount 0.000001
							if solRequiredAmountFloat < 0.000001 {
								solRequiredAmount = "0.000001"
							}

							// Buy solXEN
							result, err := utils.ExchangeSolForToken(solRequiredAmount, "solXEN")
							if err != nil {
								utils.LogMessage(moduleUI.LogView, "Error buying solXEN: "+err.Error())
							} else {
								utils.LogMessage(moduleUI.LogView, fmt.Sprintf("SOL: %s -> solXEN: %s successfully for governance purposes", solRequiredAmount, result))
							}

							// Wait for transaction to complete
							time.Sleep(60 * time.Second)
						}
					}

					// Execute token swap based on configuration
					if solBalance < config.SOLPerHarvest+0.000005 {
						utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Insufficient SOL balance %f. Minimum required: %.9f SOL", solBalance, config.SOLPerHarvest+0.000005))
						break counterdownLoop
					}

					// Check if SOL balance is sufficient
					if solBalance < 0.000006 {
						utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Insufficient SOL balance %f. Minimum required: 0.000006 SOL", solBalance))
						break counterdownLoop
					}

					// Ensure minimum SOL amount 0.000001
					if config.SOLPerHarvest < 0.000001 {
						config.SOLPerHarvest = 0.000001
					}

					result, err := utils.ExchangeSolForToken(strconv.FormatFloat(config.SOLPerHarvest, 'f', -1, 64), config.TokenToHarvest)
					if err != nil {
						utils.LogMessage(moduleUI.LogView, "Error: "+err.Error())
					} else {
						utils.LogMessage(moduleUI.LogView, fmt.Sprintf("SOL: %v -> %s: %s successfully", config.SOLPerHarvest, config.TokenToHarvest, result))

						// Update wallet info after 60 seconds
						go func() {
							time.Sleep(60 * time.Second)
							UpdateWalletInfo(app, walletInfoView)
						}()
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
	manualHarvestForm := tview.NewForm()
	manualHarvestForm.SetBorder(true).
		SetTitle("Manual Harvest").
		SetTitleAlign(tview.AlignLeft)

	var tokenAmountText *tview.TextView
	var selectedToken string
	solAmount := "0.1"
	tokenIndex := 0 // Default to solXEN
	selectedToken = tokenOptions[tokenIndex]

	updateTokenAmount := func() {
		go func() {
			if solAmount != "" && !regexp.MustCompile(`^[0.]*$`).MatchString(solAmount) {
				tokenAmount, err := utils.GetTokenExchangeAmount(solAmount, selectedToken)
				if err != nil {
					utils.LogMessage(moduleUI.LogView, "Error calculating token amount: "+err.Error())
				} else {
					tokenAmountText.SetText("Amount(Est.): \n" + tokenAmount + " " + selectedToken)
				}
			} else {
				tokenAmountText.SetText("Amount(Est.): \n-")
			}
		}()
	}

	// 1. SOL Amount input field
	manualHarvestForm.AddInputField("SOL Amount", solAmount, 10, func(textToCheck string, lastChar rune) bool {
		// Allow digits
		if lastChar >= '0' && lastChar <= '9' {
			return true
		}
		// Allow one decimal point, but not as the first character
		if lastChar == '.' && strings.Contains(textToCheck, ".") && len(textToCheck) > 0 {
			return true
		}
		return false
	}, func(text string) {
		solAmount = text
		// Calculate solXEN amount when SOL amount changes
		updateTokenAmount()
	})

	// Create the token dropdown
	tokenDropdown := tview.NewDropDown().
		SetLabel("Token to Harvest").
		SetOptions(tokenOptions, func(text string, index int) {
			selectedToken = text
			updateTokenAmount()
		}).
		SetCurrentOption(tokenIndex)

	manualHarvestForm.AddFormItem(tokenDropdown)

	// 3. Token amount field
	tokenAmountText = tview.NewTextView()
	tokenAmountText.SetText("Amount(Est.)")

	manualHarvestForm.AddFormItem(tokenAmountText)

	// Initial token amount calculation
	updateTokenAmount()

	// 3. Swap button
	manualHarvestForm.AddButton("Manual Harvest", func() {
		utils.LogMessage(moduleUI.LogView, fmt.Sprintf("Swapping %s SOL for solXEN", solAmount))

		result, err := utils.ExchangeSolForToken(solAmount, selectedToken)
		if err != nil {
			// Handle error
			utils.LogMessage(moduleUI.LogView, "Error: "+err.Error())
		} else {
			tokenAmountText.SetText("Amount(Est.): \n" + result + " " + selectedToken)
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
