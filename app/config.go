package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	GamePaksDir                      string `json:"GamePaksDir,omitempty"`
	UsmapPath                        string `json:"UsmapPath,omitempty"`
	AesKey                           string `json:"AesKey,omitempty"`
	OutputExtractionDir              string `json:"OutputExtractionDir,omitempty"`
	PreviewCommand                   bool   `json:"PreviewCommand,omitempty"`
	EnableAdvancedExtractIoStoreArgs bool   `json:"EnableAdvancedExtractIoStoreArgs,omitempty"`
	ToolVersion                      string `json:"ToolVersion,omitempty"`
}

func configPath() string {
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func loadConfig() Config {
	var cfg Config
	data, err := os.ReadFile(configPath())
	if err != nil {
		fmt.Println("[debug] loadConfig read skipped:", err)
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("[debug] loadConfig unmarshal failed:", err)
	}
	return cfg
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}
