package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type UnmineableInfo struct {
	Uuid             string
	Balance          string
	PaymentThreshold string
	AutoPay          bool
	Coin             string
	Past24h          string
	Past7d           string
	Past30d          string
}

var (
	unmineableInfo UnmineableInfo
)

// GetUnmineableInfo fetches information from the Unmineable API
func GetUnmineableInfo(publicKey string, coin string) (*UnmineableInfo, error) {
	url := fmt.Sprintf("https://api.unminable.com/v4/address/%s?coin=%s", publicKey, coin)

	resp, err := http.Get(url)
	if err != nil {
		LogToFile(fmt.Sprintf("HTTP GET request failed: %v", err))
		return nil, fmt.Errorf("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to read response body: %v", err))
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var unmineableResp struct {
		// Add fields based on the actual API response structure
		// This is a placeholder and should be adjusted
		Data struct {
			Uuid    string `json:"uuid"`
			AutoPay bool   `json:"auto"`
			// Add other fields as needed
		} `json:"data"`
	}
	err = json.Unmarshal(body, &unmineableResp)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to unmarshal JSON: %v", err))
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Log the unmineableResp to the debug file
	LogToFile(fmt.Sprintf("Unmineable Response: %+v", unmineableResp))

	unmineableInfo = UnmineableInfo{
		Uuid:    unmineableResp.Data.Uuid,
		AutoPay: unmineableResp.Data.AutoPay,
	}

	// Fetch additional information
	url = fmt.Sprintf("https://api.unminable.com/v4/account/%s/stats", unmineableInfo.Uuid)
	resp, err = http.Get(url)
	if err != nil {
		LogToFile(fmt.Sprintf("HTTP GET request failed: %v", err))
		return nil, fmt.Errorf("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to read response body: %v", err))
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var unmineableStatsResp struct {
		Data struct {
			Balance          string `json:"balance"`
			PaymentThreshold string `json:"payment_threshold"`
			Coin             string `json:"coin"`
			Rewarded         struct {
				Past24h string `json:"past_24h"`
				Past7d  string `json:"past_7d"`
				Past30d string `json:"past_30d"`
			} `json:"rewarded"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &unmineableStatsResp)
	if err != nil {
		LogToFile(fmt.Sprintf("Failed to unmarshal JSON: %v", err))
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	unmineableInfo.Balance = unmineableStatsResp.Data.Balance
	unmineableInfo.PaymentThreshold = unmineableStatsResp.Data.PaymentThreshold
	unmineableInfo.Coin = unmineableStatsResp.Data.Coin
	unmineableInfo.Past24h = unmineableStatsResp.Data.Rewarded.Past24h
	unmineableInfo.Past7d = unmineableStatsResp.Data.Rewarded.Past7d
	unmineableInfo.Past30d = unmineableStatsResp.Data.Rewarded.Past30d

	LogToFile(fmt.Sprintf("Fetched Unmineable Info: %+v", unmineableInfo))

	return &unmineableInfo, nil
}
