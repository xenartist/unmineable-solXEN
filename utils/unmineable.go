package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// UnmineableResponse represents the structure of the API response
type UnmineableResponse struct {
	// Add fields based on the actual API response structure
	// This is a placeholder and should be adjusted
	Data struct {
		Balance          string `json:"balance"`
		PaymentThreshold string `json:"payment_threshold"`
		AutoPay          bool   `json:"auto"`
		Network          string `json:"network"`
		// Add other fields as needed
	} `json:"data"`
}

// GetUnmineableInfo fetches information from the Unmineable API
func GetUnmineableInfo(publicKey string, coin string) (*UnmineableResponse, error) {
	url := fmt.Sprintf("https://api.unminable.com/v4/address/%s?coin=%s", publicKey, coin)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var unmineableResp UnmineableResponse
	err = json.Unmarshal(body, &unmineableResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Log the unmineableResp to the debug file
	logMessage := fmt.Sprintf("Unmineable Response: %+v", unmineableResp)
	LogToFile(logMessage)

	return &unmineableResp, nil
}
