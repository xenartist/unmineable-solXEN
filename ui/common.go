package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"xoon/utils"

	"github.com/rivo/tview"
)

const WALLET_STRING = "Solana Wallet"
const SOLXEN_CPU_MINER_STRING = "solXEN Miner (CPU)"
const SOLXEN_GPU_MINER_STRING = "solXEN Miner (GPU)"
const TOKEN_HARVEST_STRING = "Token Harvest"

var ModuleNames = []string{WALLET_STRING, SOLXEN_CPU_MINER_STRING, SOLXEN_GPU_MINER_STRING, TOKEN_HARVEST_STRING}

type ModuleUI struct {
	DashboardFlex *tview.Flex
	LogView       *tview.TextView
	ConfigFlex    *tview.Flex
}

var (
	dashboardFlexInstance *tview.Flex
	dashboardMutex        sync.Mutex
)

var walletInfoView *tview.TextView

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

			// Get solXEN equivalent
			solXENAmount, err := utils.GetTokenExchangeAmount(info.Balance, utils.SolXEN)
			if err != nil {
				utils.LogToFile(fmt.Sprintf("Error getting solXEN amount: %v", err))
				solXENAmount = "0" // Set to 0 if there's an error
			}

			// First line
			lin0 := "MINING STATS:"
			line1 := fmt.Sprintf("Pending: %s %s (%s solXEN) | AutoPay: %s | PayOn: %s %s",
				info.Balance, info.Coin,
				solXENAmount,
				autoPay,
				info.PaymentThreshold, info.Coin)

			// Second line - calculate rewards in solXEN
			solXEN24h, _ := utils.GetTokenExchangeAmount(info.Past24h, utils.SolXEN)
			solXEN7d, _ := utils.GetTokenExchangeAmount(info.Past7d, utils.SolXEN)
			solXEN30d, _ := utils.GetTokenExchangeAmount(info.Past30d, utils.SolXEN)

			line2 := fmt.Sprintf("Rewards: 24h: %s solXEN | 7d: %s solXEN | 30d: %s solXEN",
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
	walletInfoView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Initial update
	UpdateWalletInfo(app, walletInfoView)

	// Set up a ticker to update every 60 minutes
	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		for range ticker.C {
			UpdateWalletInfo(app, walletInfoView)
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
	case TOKEN_HARVEST_STRING:
		return CreateTokenHarvestConfigFlex(app, logView)

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
				AddItem(configFlex, 0, 3, false),
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

// Function to update wallet info
func UpdateWalletInfo(app *tview.Application, walletInfoView *tview.TextView) {
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

		// Define the desired order of tokens
		tokenOrder := []string{"solXEN", "xencat", "ORE"}

		// Create a map for quick lookup of balances
		balanceMap := make(map[string]utils.TokenBalance)
		for _, balance := range balances {
			balanceMap[balance.Symbol] = balance
		}

		// Write balances in the specified order
		for _, symbol := range tokenOrder {
			if balance, exists := balanceMap[symbol]; exists {
				infoText.WriteString(fmt.Sprintf(" | %.6f %s ", balance.Balance, balance.Symbol))
			} else {
				// If the token doesn't exist in the balances, display zero
				infoText.WriteString(fmt.Sprintf(" | 0.000000 %s ", symbol))
			}
		}

		app.QueueUpdateDraw(func() {
			walletInfoView.SetText(infoText.String())
		})
	}()
}
