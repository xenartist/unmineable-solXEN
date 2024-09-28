package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	SolXEN          = "solXEN"
	OGSolXEN        = "OG solXEN"
	JupiterQuoteURL = "https://quote-api.jup.ag/v6/quote"
	JupiterSwapURL  = "https://quote-api.jup.ag/v6/swap"
	JupiterPriceURL = "https://price.jup.ag/v6/price"
	SOLMint         = "So11111111111111111111111111111111111111112"
)

type QuoteResponse struct {
	InputMint            string  `json:"inputMint"`
	InAmount             string  `json:"inAmount"`
	OutputMint           string  `json:"outputMint"`
	OutAmount            string  `json:"outAmount"`
	OtherAmountThreshold string  `json:"otherAmountThreshold"`
	SwapMode             string  `json:"swapMode"`
	SlippageBps          int     `json:"slippageBps"`
	PriceImpactPct       string  `json:"priceImpactPct"`
	RoutePlan            []Route `json:"routePlan"`
}

type Route struct {
	SwapInfo SwapInfo `json:"swapInfo"`
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
	QuoteResponse     QuoteResponse `json:"quoteResponse"`
	UserPublicKey     string        `json:"userPublicKey"`
	WrapUnwrapSOL     bool          `json:"wrapUnwrapSOL"`
	UseSharedAccounts bool          `json:"useSharedAccounts"`
}

type SwapResponse struct {
	SwapTransaction string `json:"swapTransaction"`
}

var (
	tokenAddresses = map[string]string{
		SolXEN: "6f8deE148nynnSiWshA9vLydEbJGpDeKh5G4PRgjmzG7",
		// OGSolXEN: "EEqrab5tdnVdZv7a4AUAvGehDAtM8gWd7szwfyYbmGkM",
	}
	solXENPrice float64
	// ogSolXENPrice float64
	priceMutex sync.RWMutex
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
}

func fetchPrice(tokenName string) float64 {
	LogToFile(fmt.Sprintf("Fetching price for %s", tokenName))

	tokenAddress, ok := tokenAddresses[tokenName]
	if !ok {
		LogToFile(fmt.Sprintf("Unknown token: %s", tokenName))
		return 0
	}

	apiURL := fmt.Sprintf("%s?ids=SOL&vsToken=%s", JupiterPriceURL, tokenAddress)

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
			Price float64 `json:"price"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		LogToFile(fmt.Sprintf("Failed to parse JSON response: %v", err))
		return 0
	}

	priceData, ok := priceResponse.Data["SOL"]
	if !ok {
		LogToFile("Price not found in response")
		return 0
	}

	LogToFile(fmt.Sprintf("Fetched price for %s: %f", tokenName, priceData.Price))
	return priceData.Price
}

// GetTokenExchangeAmount queries the amount of specified token that can be exchanged for a given amount of SOL
func GetTokenExchangeAmount(solAmount float64, tokenName string) (float64, error) {
	priceMutex.RLock()
	defer priceMutex.RUnlock()

	var price float64
	switch tokenName {
	case SolXEN:
		price = solXENPrice
	// case OGSolXEN:
	// 	price = ogSolXENPrice
	default:
		return 0, fmt.Errorf("Unknown token: %s", tokenName)
	}

	if price == 0 {
		return 0, errors.New("Price not available")
	}

	result := solAmount * price
	LogToFile(fmt.Sprintf("Calculated result for %s: %f", tokenName, result))
	return result, nil
}

func ExchangeSolForToken(solAmount float64, tokenName string) (float64, error) {
	// Step 1: Get the token mint address
	tokenMint, ok := tokenAddresses[tokenName]
	if !ok {
		return 0, fmt.Errorf("unknown token: %s", tokenName)
	}

	// Step 2: Get a quote
	quoteResp, err := getQuote(SOLMint, tokenMint, fmt.Sprintf("%.9f", solAmount))
	if err != nil {
		return 0, fmt.Errorf("failed to get quote: %v", err)
	}

	// Step 3: Execute the swap
	swapResp, err := executeSwap(quoteResp, GetGlobalPublicKey())
	if err != nil {
		return 0, fmt.Errorf("failed to execute swap: %v", err)
	}

	// Step 4: Parse the output amount
	outAmount, err := strconv.ParseFloat(quoteResp.OutAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse out amount: %v", err)
	}

	// Log the transaction for the user to sign and send
	LogToFile(fmt.Sprintf("Swap transaction: %s", swapResp.SwapTransaction))

	return outAmount, nil
}

func getQuote(inputMint, outputMint, inAmount string) (*QuoteResponse, error) {
	url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s&amount=%s&slippageBps=50",
		JupiterQuoteURL, inputMint, outputMint, inAmount)

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

	return &quoteResp, nil
}

func executeSwap(quote *QuoteResponse, userPublicKey string) (*SwapResponse, error) {
	swapRequest := SwapRequest{
		QuoteResponse:     *quote,
		UserPublicKey:     userPublicKey,
		WrapUnwrapSOL:     true,
		UseSharedAccounts: true,
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
		return nil, err
	}

	return &swapResp, nil
}
