package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the configuration for the Bitcask database
type Config struct {
	DataDir   string `json:"data_dir"`
	DebugMode bool   `json:"debug_mode"`
}

// BitcaskConfig manages configuration for the Bitcask database
type BitcaskConfig struct {
	config Config
}

// NewConfig creates a new configuration manager
func NewConfig() *BitcaskConfig {
	bc := &BitcaskConfig{}
	bc.loadConfig()
	return bc
}

// GetDefaultDataDir returns the platform-specific default data directory
func (bc *BitcaskConfig) GetDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		return filepath.Join(homeDir, "Library", "Application Support", "gobitcask")
	case "linux":
		return "/var/lib/gobitcask"
	case "windows":
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		return filepath.Join(programData, "gobitcask")
	default:
		return filepath.Join(homeDir, ".gobitcask")
	}
}

// GetConfigDir returns the platform-specific configuration directory
func (bc *BitcaskConfig) GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		return filepath.Join(homeDir, "Library", "Preferences", "gobitcask")
	case "linux":
		return filepath.Join(homeDir, ".config", "gobitcask")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "gobitcask")
	default:
		return filepath.Join(homeDir, ".gobitcask")
	}
}

// GetConfigFile returns the path to the configuration file
func (bc *BitcaskConfig) GetConfigFile() string {
	return filepath.Join(bc.GetConfigDir(), "config.json")
}

// loadConfig loads configuration from file or creates default
func (bc *BitcaskConfig) loadConfig() {
	configFile := bc.GetConfigFile()

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default configuration
		bc.config = Config{
			DataDir:   bc.GetDefaultDataDir(),
			DebugMode: false,
		}
		bc.saveConfig()
		return
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		// Use default configuration if file can't be read
		bc.config = Config{
			DataDir:   bc.GetDefaultDataDir(),
			DebugMode: false,
		}
		return
	}

	if err := json.Unmarshal(data, &bc.config); err != nil {
		// Use default configuration if JSON is invalid
		bc.config = Config{
			DataDir:   bc.GetDefaultDataDir(),
			DebugMode: false,
		}
	}
}

// saveConfig saves the current configuration to file
func (bc *BitcaskConfig) saveConfig() error {
	configDir := bc.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(bc.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(bc.GetConfigFile(), data, 0644)
}

// GetDataDir returns the data directory, with optional override
func (bc *BitcaskConfig) GetDataDir(overrideDir string) string {
	if overrideDir != "" {
		return overrideDir
	}
	return bc.config.DataDir
}

// GetDebugMode returns the debug mode setting
func (bc *BitcaskConfig) GetDebugMode() bool {
	return bc.config.DebugMode
}

// SetDataDir sets the data directory in configuration
func (bc *BitcaskConfig) SetDataDir(dataDir string) error {
	bc.config.DataDir = dataDir
	return bc.saveConfig()
}

// SetDebugMode sets the debug mode in configuration
func (bc *BitcaskConfig) SetDebugMode(debugMode bool) error {
	bc.config.DebugMode = debugMode
	return bc.saveConfig()
}

// Global configuration instance
var GlobalConfig = NewConfig()
