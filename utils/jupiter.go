package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
)

const (
	SolXEN          = "solXEN"
	OGSolXEN        = "OG solXEN"
	xencat          = "xencat"
	PV              = "PV"
	ORE             = "ORE"
	JupiterQuoteURL = "https://quote-api.jup.ag/v6/quote"
	JupiterSwapURL  = "https://quote-api.jup.ag/v6/swap"
	JupiterPriceURL = "https://api.jup.ag/price/v2"
	SOLMint         = "So11111111111111111111111111111111111111112"
)

type RoutePlanItem struct {
	SwapInfo SwapInfo `json:"swapInfo"`
	Percent  int      `json:"percent"`
}

type QuoteResponse struct {
	InputMint            string          `json:"inputMint"`
	InAmount             string          `json:"inAmount"`
	OutputMint           string          `json:"outputMint"`
	OutAmount            string          `json:"outAmount"`
	OtherAmountThreshold string          `json:"otherAmountThreshold"`
	SwapMode             string          `json:"swapMode"`
	SlippageBps          int             `json:"slippageBps"`
	PriceImpactPct       string          `json:"priceImpactPct"`
	RoutePlan            []RoutePlanItem `json:"routePlan"`
}

type SwapInfo struct {
	AmmKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

type SwapRequest struct {
	QuoteResponse             QuoteResponse `json:"quoteResponse"`
	UserPublicKey             string        `json:"userPublicKey"`
	WrapAndUnwrapSOL          bool          `json:"wrapAndUnwrapSOL"`
	PrioritizationFeeLamports string        `json:"prioritizationFeeLamports"`
}

type SwapResponse struct {
	SwapTransaction string `json:"swapTransaction"`
}

var (
	tokenAddresses = map[string]string{
		SolXEN: "6f8deE148nynnSiWshA9vLydEbJGpDeKh5G4PRgjmzG7",
		// OGSolXEN: "EEqrab5tdnVdZv7a4AUAvGehDAtM8gWd7szwfyYbmGkM",
		xencat: "7UN8WkBumTUCofVPXCPjNWQ6msQhzrg9tFQRP48Nmw5V",
		PV:     "5px3a5LWR6CmiYX3ktpNnGYiEypfDdemRd74GDYbsJ2H",
		ORE:    "oreoU2P8bN6jkk3jbaiVxYnG1dCXcYxwhwyK9jSybcp",
	}
	solXENPrice float64 = 0
	// ogSolXENPrice float64
	xencatPrice float64 = 0
	pvPrice     float64 = 0
	orePrice    float64 = 0
	priceMutex  sync.RWMutex
)

func InitJupiter() {
	updatePrices()

	// Update prices every 30 minutes
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			updatePrices()
		}
	}()
}

func updatePrices() {
	solXENPrice = fetchPrice(SolXEN)
	// ogSolXENPrice = fetchPrice(OGSolXEN)
	xencatPrice = fetchPrice(xencat)
	pvPrice = fetchPrice(PV)
	orePrice = fetchPrice(ORE)
}

func fetchPrice(tokenName string) float64 {
	LogToFile(fmt.Sprintf("Fetching price for %s", tokenName))

	tokenAddress, ok := tokenAddresses[tokenName]
	if !ok {
		LogToFile(fmt.Sprintf("Unknown token: %s", tokenName))
		return 0
	}

	apiURL := fmt.Sprintf("%s?ids=%s&vsToken=%s", JupiterPriceURL, SOLMint, tokenAddress)

	resp, err := http.Get(apiURL)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to get price from Jupiter API: %v", err))
		return 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to read response body: %v", err))
		return 0
	}

	var priceResponse struct {
		Data map[string]struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Price string `json:"price"`
		} `json:"data"`
		TimeTaken float64 `json:"timeTaken"`
	}
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		LogToFile(fmt.Sprintf("Failed to parse JSON response: %v", err))
		return 0
	}

	priceData, ok := priceResponse.Data[SOLMint]
	if !ok {
		LogToFile("Price not found in response")
		return 0
	}

	price, err := strconv.ParseFloat(priceData.Price, 64)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to parse price string: %v", err))
		return 0
	}

	LogToFile(fmt.Sprintf("Fetched price for %s: %f", tokenName, price))
	return price
}

// GetTokenExchangeAmount queries the amount of specified token that can be exchanged for a given amount of SOL
func GetTokenExchangeAmount(solAmount string, tokenName string) (string, error) {
	priceMutex.RLock()
	defer priceMutex.RUnlock()

	solAmountFloat, err := strconv.ParseFloat(solAmount, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse SOL amount: %w", err)
	}

	var price float64
	switch tokenName {
	case SolXEN:
		price = solXENPrice
	// case OGSolXEN:
	// 	price = ogSolXENPrice
	case xencat:
		price = xencatPrice
	case PV:
		price = pvPrice
	case ORE:
		price = orePrice
	default:
		return "", fmt.Errorf("Unknown token: %s", tokenName)
	}

	if price == 0 {
		return "", errors.New("Price not available")
	}

	result := solAmountFloat * price

	formattedResult := fmt.Sprintf("%.6f", result)
	LogToFile(fmt.Sprintf("Calculated result for %s: %s", tokenName, formattedResult))

	return formattedResult, nil
}

func GetSolExchangeAmount(tokenAmount string, tokenName string) (string, error) {
	priceMutex.RLock()
	defer priceMutex.RUnlock()

	tokenAmountFloat, err := strconv.ParseFloat(tokenAmount, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse token amount: %w", err)
	}

	var price float64
	switch tokenName {
	case SolXEN:
		price = solXENPrice
	case xencat:
		price = xencatPrice
	case PV:
		price = pvPrice
	case ORE:
		price = orePrice
	default:
		return "", fmt.Errorf("Unknown token: %s", tokenName)
	}

	if price == 0 {
		return "", errors.New("Price not available")
	}

	result := tokenAmountFloat / price

	formattedResult := fmt.Sprintf("%.9f", result) // 9 decimal places for SOL
	LogToFile(fmt.Sprintf("Calculated SOL amount for %s %s: %s", tokenAmount, tokenName, formattedResult))

	return formattedResult, nil
}

func ExchangeSolForToken(solAmount string, tokenName string) (string, error) {
	// Step 1: Get the token mint address
	tokenMint, ok := tokenAddresses[tokenName]
	if !ok {
		return "", fmt.Errorf("unknown token: %s", tokenName)
	}

	// Step 2: Get a quote
	quoteResp, err := getQuote(SOLMint, tokenMint, solAmount)
	if err != nil {
		return "", fmt.Errorf("failed to get quote: %v", err)
	}

	// Step 3: Execute the swap
	swapResp, err := executeSwap(quoteResp, GetGlobalPublicKey())
	if err != nil {
		return "", fmt.Errorf("failed to execute swap: %v", err)
	}

	// Step 4: Parse the output amount
	outAmount, err := parseAmount(quoteResp.OutAmount)
	if err != nil {
		return "", fmt.Errorf("failed to parse out amount: %v", err)
	}

	// Step 5: Sign and send the transaction
	err = signAndSendTransaction(swapResp.SwapTransaction, getPrivateKey())
	if err != nil {
		return "", fmt.Errorf("failed to sign and send transaction: %v", err)
	}

	// Log the transaction for the user to sign and send
	LogToFile(fmt.Sprintf("Swap transaction: %s", swapResp.SwapTransaction))

	return strconv.FormatFloat(outAmount/1_000_000, 'f', -1, 64), nil
}

// Helper function to adjust the amount based on token decimals
func adjustAmountForDecimals(amount string) (string, error) {
	// Convert string to decimal
	decimalAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return "", fmt.Errorf("failed to convert amount to decimal: %w", err)
	}

	// Multiply by 10^9 (1_000_000_000)
	adjustedAmount := decimalAmount.Mul(decimal.NewFromInt(1_000_000_000))

	LogToFile(fmt.Sprintf("Original amount: %s, Adjusted amount: %s", amount, adjustedAmount.String()))

	// Return the string representation of the adjusted decimal
	return adjustedAmount.String(), nil
}

func getQuote(inputMint string, outputMint string, inAmount string) (*QuoteResponse, error) {

	// Adjust inAmount
	adjustedAmount, err := adjustAmountForDecimals(inAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust input amount: %w", err)
	}

	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%s&slippageBps=50",
		JupiterQuoteURL, inputMint, outputMint, adjustedAmount)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quoteResp QuoteResponse
	err = json.Unmarshal(body, &quoteResp)
	if err != nil {
		return nil, err
	}

	// Format the JSON for logging
	formattedJSON, err := json.MarshalIndent(quoteResp, "", "  ")
	if err != nil {
		LogToFile(fmt.Sprintf("Error formatting JSON: %v\nRaw response:\n%s", err, string(body)))
	} else {
		LogToFile(fmt.Sprintf("Quote Response:\n%s", string(formattedJSON)))
	}

	return &quoteResp, nil
}

func executeSwap(quote *QuoteResponse, userPublicKey string) (*SwapResponse, error) {
	swapRequest := SwapRequest{
		QuoteResponse:             *quote,
		UserPublicKey:             userPublicKey,
		WrapAndUnwrapSOL:          true,
		PrioritizationFeeLamports: "auto",
	}

	jsonData, err := json.Marshal(swapRequest)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(JupiterSwapURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var swapResp SwapResponse
	err = json.Unmarshal(body, &swapResp)
	if err != nil {
		// Log the raw response body
		LogToFile(fmt.Sprintf("Raw response body:\n%s", string(body)))

		// Check if the response contains an error message
		var errorResp struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errorResp); jsonErr == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("API error: %s", errorResp.Error)
		}

		// If it's not a known error format, return the original error
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	LogToFile(fmt.Sprintf("Swap Response:\n%s", string(body)))

	return &swapResp, nil
}

func signAndSendTransaction(transaction string, privateKey string) error {
	// 1. Decode the transaction data
	decodedTransaction, err := base64.StdEncoding.DecodeString(transaction)
	if err != nil {
		return fmt.Errorf("failed to decode transaction: %v", err)
	}

	LogToFile(fmt.Sprintf("Decoded transaction length: %d bytes", len(decodedTransaction)))

	if len(decodedTransaction) == 0 {
		return fmt.Errorf("decoded transaction is empty")
	}

	// 2. Parse the transaction
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(decodedTransaction))
	if err != nil {
		return fmt.Errorf("failed to parse transaction (length: %d bytes): %w", len(decodedTransaction), err)
	}

	// 3. Sign the transaction using the private key
	kp, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}
	tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(kp.PublicKey()) {
			return &kp
		}
		return nil
	})

	// 4. Serialize the signed transaction
	signedTx, err := tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to serialize signed transaction: %v", err)
	}

	// 5. Send the transaction to the Solana network
	client := rpc.New("https://api.mainnet-beta.solana.com")
	// Convert signedTx to *solana.Transaction
	tx, err = solana.TransactionFromDecoder(bin.NewBinDecoder(signedTx))
	if err != nil {
		return fmt.Errorf("failed to decode transaction: %v", err)
	}
	sig, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	LogToFile(fmt.Sprintf("Transaction sent: %s", sig))
	return nil
}

func parseAmount(amountStr string) (float64, error) {
	// Check if the string is empty
	if amountStr == "" {
		return 0, fmt.Errorf("amount string is empty")
	}

	// Try to parse the string to float64
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount: %v", err)
	}

	return amount, nil
}
