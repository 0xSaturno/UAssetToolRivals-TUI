package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	GamePaksDir                    string `json:"GamePaksDir,omitempty"`
	UsmapPath                      string `json:"UsmapPath,omitempty"`
	AesKey                         string `json:"AesKey,omitempty"`
	OutputExtractionDir            string `json:"OutputExtractionDir,omitempty"`
	PreviewCommand                 bool   `json:"PreviewCommand,omitempty"`
	EnableAdvancedExtractIoStoreArgs bool   `json:"EnableAdvancedExtractIoStoreArgs,omitempty"`
	ToolVersion                     string `json:"ToolVersion,omitempty"`
}

func configPath() string {
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func loadConfig() Config {
	var cfg Config
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}
