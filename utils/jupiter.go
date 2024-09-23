package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	SolXEN   = "solXEN"
	OGSolXEN = "OG solXEN"
)

var tokenAddresses = map[string]string{
	SolXEN:   "6f8deE148nynnSiWshA9vLydEbJGpDeKh5G4PRgjmzG7",
	OGSolXEN: "EEqrab5tdnVdZv7a4AUAvGehDAtM8gWd7szwfyYbmGkM",
}

// GetTokenExchangeAmount queries the amount of specified token that can be exchanged for a given amount of SOL
func GetTokenExchangeAmount(solAmount float64, tokenName string) (float64, error) {
	LogToFile(fmt.Sprintf("Starting GetTokenExchangeAmount with solAmount: %f, tokenName: %s", solAmount, tokenName))

	tokenAddress, ok := tokenAddresses[tokenName]
	if !ok {
		errMsg := fmt.Sprintf("Unknown token: %s", tokenName)
		LogToFile(errMsg)
		return 0, fmt.Errorf(errMsg)
	}

	// Create the API request URL
	apiURL := fmt.Sprintf("https://price.jup.ag/v6/price?ids=SOL&vsToken=%s", tokenAddress)

	// Make the HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get price from Jupiter API: %v", err)
		LogToFile(errMsg)
		return 0, fmt.Errorf(errMsg)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read response body: %v", err)
		LogToFile(errMsg)
		return 0, fmt.Errorf(errMsg)
	}

	// Parse the JSON response
	var priceResponse struct {
		Data map[string]struct {
			Price float64 `json:"price"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		errMsg := fmt.Sprintf("Failed to parse JSON response: %v", err)
		LogToFile(errMsg)
		return 0, fmt.Errorf(errMsg)
	}

	// Get the price from the response
	priceData, ok := priceResponse.Data["SOL"]
	if !ok {
		errMsg := "Price not found in response"
		return 0, errors.New(errMsg)
	}
	price := priceData.Price

	// Calculate the result
	result := solAmount * price

	LogToFile(fmt.Sprintf("Calculated result: %f", result))

	return result, nil
}
