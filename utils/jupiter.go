package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	SolXEN   = "solXEN"
	OGSolXEN = "OG solXEN"
)

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

	apiURL := fmt.Sprintf("https://price.jup.ag/v6/price?ids=SOL&vsToken=%s", tokenAddress)

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
