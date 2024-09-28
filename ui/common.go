package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

const WALLET_STRING = "Solana Wallet"
const SOLXEN_CPU_MINER_STRING = "solXEN Miner (CPU)"
const SOLXEN_GPU_MINER_STRING = "solXEN Miner (GPU)"

var ModuleNames = []string{WALLET_STRING, SOLXEN_CPU_MINER_STRING, SOLXEN_GPU_MINER_STRING}

type ModuleUI struct {
	DashboardFlex *tview.Flex
	LogView       *tview.TextView
	ConfigFlex    *tview.Flex
}

var (
	dashboardFlexInstance *tview.Flex
	dashboardMutex        sync.Mutex
)

func GetDashboardFlex(app *tview.Application) *tview.Flex {
	dashboardMutex.Lock()
	defer dashboardMutex.Unlock()

	if dashboardFlexInstance == nil {
		dashboardFlexInstance = createDashboardFlex(app)
	}

	return dashboardFlexInstance
}

func CreateModuleUI(name string, app *tview.Application) ModuleUI {
	logView := CreateLogView(name+" Logs", app)
	configFlex := CreateConfigFlex(name, app, logView)
	dashboardFlex := GetDashboardFlex(app)
	return ModuleUI{
		DashboardFlex: dashboardFlex,
		LogView:       logView,
		ConfigFlex:    configFlex,
	}
}

func createDashboardFlex(app *tview.Application) *tview.Flex {
	dashboardFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Create a new text view for Unmineable info
	unmineableInfoView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Function to update Unmineable info
	updateUnmineableInfo := func() {
		// Check if GLOBAL_PUBLIC_KEY is empty
		if utils.GetGlobalPublicKey() == "" {
			return
		}

		go func() {
			info, err := utils.GetUnmineableInfo(utils.GetGlobalPublicKey(), "SOL")
			if err != nil {
				app.QueueUpdateDraw(func() {
					unmineableInfoView.SetText(fmt.Sprintf("Error: %v", err))
				})
				return
			}

			autoPay := "Off"
			if info.AutoPay {
				autoPay = "On"
			}

			// Convert Balance string to float64
			balance, err := strconv.ParseFloat(info.Balance, 64)
			if err != nil {
				utils.LogToFile(fmt.Sprintf("Error parsing balance: %v", err))
				balance = 0 // Set to 0 if parsing fails
			}

			// Get solXEN equivalent
			solXENAmount, err := utils.GetTokenExchangeAmount(balance, utils.SolXEN)
			if err != nil {
				utils.LogToFile(fmt.Sprintf("Error getting solXEN amount: %v", err))
				solXENAmount = 0 // Set to 0 if there's an error
			}

			// First line
			lin0 := "MINING STATS:"
			line1 := fmt.Sprintf("Pending: %s %s (%.2f solXEN) | AutoPay: %s | PayOn: %s %s",
				info.Balance, info.Coin,
				solXENAmount,
				autoPay,
				info.PaymentThreshold, info.Coin)

			// Second line - calculate rewards in solXEN
			past24h, _ := strconv.ParseFloat(info.Past24h, 64)
			past7d, _ := strconv.ParseFloat(info.Past7d, 64)
			past30d, _ := strconv.ParseFloat(info.Past30d, 64)

			solXEN24h, _ := utils.GetTokenExchangeAmount(past24h, utils.SolXEN)
			solXEN7d, _ := utils.GetTokenExchangeAmount(past7d, utils.SolXEN)
			solXEN30d, _ := utils.GetTokenExchangeAmount(past30d, utils.SolXEN)

			line2 := fmt.Sprintf("Rewards: 24h: %.2f solXEN | 7d: %.2f solXEN | 30d: %.2f solXEN",
				solXEN24h, solXEN7d, solXEN30d)

			// Combine both lines
			infoText := lin0 + "\n" + line1 + "\n" + line2

			app.QueueUpdateDraw(func() {
				unmineableInfoView.SetText(infoText)
			})
		}()
	}

	// Initial update
	updateUnmineableInfo()

	// Set up a ticker to update every 60 minutes
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			updateUnmineableInfo()
		}
	}()

	// Create a new text view for wallet info
	walletInfoView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Function to update wallet info
	updateWalletInfo := func() {
		if utils.GetGlobalPublicKey() == "" {
			return
		}

		go func() {
			var infoText strings.Builder
			infoText.WriteString("\n")
			infoText.WriteString("WALLET BALANCE: \n")

			// Get SOL balance
			solBalance, err := utils.GetSOLBalance(utils.GetGlobalPublicKey())
			if err != nil {
				app.QueueUpdateDraw(func() {
					walletInfoView.SetText(fmt.Sprintf("Error fetching SOL balance: %v", err))
				})
				return
			}
			infoText.WriteString(fmt.Sprintf("%.6f SOL", solBalance))

			// Get other token balances
			balances, err := utils.GetWalletTokenBalances(utils.GetGlobalPublicKey())
			if err != nil {
				app.QueueUpdateDraw(func() {
					walletInfoView.SetText(fmt.Sprintf("Error fetching wallet balances: %v", err))
				})
				return
			}

			for _, balance := range balances {
				infoText.WriteString(fmt.Sprintf(" | %.6f %s ", balance.Balance, balance.Symbol))
			}

			app.QueueUpdateDraw(func() {
				walletInfoView.SetText(infoText.String())
			})
		}()
	}

	// Initial update
	updateWalletInfo()

	// Set up a ticker to update every 60 minutes
	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		for range ticker.C {
			updateWalletInfo()
		}
	}()

	// Add the Unmineable info view to the flex
	dashboardFlex.AddItem(unmineableInfoView, 0, 1, false)
	dashboardFlex.AddItem(walletInfoView, 0, 1, false)

	dashboardFlex.SetBorder(true).SetTitle("Dashboard")
	return dashboardFlex
}

func CreateLogView(title string, app *tview.Application) *tview.TextView {
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	logView.SetBorder(true).SetTitle(title)
	return logView
}

func CreateConfigFlex(title string, app *tview.Application, logView *tview.TextView) *tview.Flex {

	switch title {
	case WALLET_STRING:
		return CreateWalletConfigFlex(app, logView)
	case SOLXEN_CPU_MINER_STRING:
		return CreateSolXENCPUConfigFlex(app, logView)
	case SOLXEN_GPU_MINER_STRING:
		return CreateSolXENGPUConfigFlex(app, logView)

	default:
		return createDefaultConfigFlex(title, app, logView)

	}
}

func createDefaultConfigFlex(title string, app *tview.Application, logView *tview.TextView) *tview.Flex {
	configFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	configFlex.SetBorder(true).SetTitle(title + " Config")
	return configFlex
}

func CreateSwitchViewFunc(rightFlex *tview.Flex, mainMenu *tview.List) func(*tview.Flex, *tview.Flex, *tview.TextView) {
	return func(dashboardFlex *tview.Flex, configFlex *tview.Flex, logView *tview.TextView) {
		rightFlex.Clear()
		rightFlex.
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(dashboardFlex, 0, 1, false).
				AddItem(configFlex, 0, 4, false),
				0, 1, false).
			AddItem(logView, 8, 1, false)
		mainMenu.SetCurrentItem(0)
	}
}

func UpdateButtonLabel(flex *tview.Flex, buttonName string, newLabel string) {
	for i := 0; i < flex.GetItemCount(); i++ {
		item := flex.GetItem(i)
		if button, ok := item.(*tview.Button); ok {
			if button.GetLabel() == buttonName {
				button.SetLabel(newLabel)
				return
			}
		}
	}
}
