package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// var GLOBAL_WORK_DIR string

type SolXENConfig struct {
	AutoHarvestActive bool    `json:"autoHarvestActive"`
	SOLPerHarvest     float64 `json:"solPerHarvest"`
	TokenToHarvest    string  `json:"tokenToHarvest"`
	HarvestInterval   string  `json:"harvestInterval"`
}

func GetExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

func XoosInit() {
	// GLOBAL_WORK_DIR = getExecutablePath()
	initFileLogger()
	initSolXENConfig()
}

func initSolXENConfig() {
	_, err := ReadSolXENConfigFile()
	if err != nil {
		// If file doesn't exist, create a default one
		defaultConfig := SolXENConfig{
			AutoHarvestActive: true,
			SOLPerHarvest:     0.01,
			TokenToHarvest:    "solXEN",
			HarvestInterval:   "1h",
		}
		err = WriteSolXENConfigFile(defaultConfig)
		if err != nil {
			LogToFile("Failed to create default config file: " + err.Error())
		}
	}
}

func ReadSolXENConfigFile() (SolXENConfig, error) {
	configPath := filepath.Join(GetExecutablePath(), "solXENconfig.json")
	file, err := os.ReadFile(configPath)
	if err != nil {
		return SolXENConfig{}, err
	}

	var config SolXENConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return SolXENConfig{}, err
	}

	return config, nil
}

func WriteSolXENConfigFile(config SolXENConfig) error {
	configPath := filepath.Join(GetExecutablePath(), "solXENconfig.json")
	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, file, 0644)
}
