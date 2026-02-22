package config

type Config struct {
	APIKey string `mapstructure:"api_key" yaml:"api_key"`
	Style  string `mapstructure:"style"   yaml:"style"`
}
