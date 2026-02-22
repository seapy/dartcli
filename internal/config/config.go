package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var DefaultConfigPath string

func init() {
	home, _ := os.UserHomeDir()
	DefaultConfigPath = filepath.Join(home, ".dartcli", "config.yaml")
}

func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		viper.AddConfigPath(filepath.Join(home, ".dartcli"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("DART")
	viper.AutomaticEnv()
	viper.BindEnv("api_key", "DART_API_KEY")

	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := filepath.Dir(DefaultConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	viper.Set("api_key", cfg.APIKey)
	viper.Set("style", cfg.Style)

	if err := viper.WriteConfigAs(DefaultConfigPath); err != nil {
		return err
	}
	return nil
}
