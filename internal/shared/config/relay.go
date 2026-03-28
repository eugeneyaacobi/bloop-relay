package config

import "github.com/spf13/viper"

type RelayToken struct {
	Name     string `mapstructure:"name"`
	Token    string `mapstructure:"token"`
	TokenEnv string `mapstructure:"token_env"`
}

type HostnameGeneration struct {
	Mode   string `mapstructure:"mode"`
	Length int    `mapstructure:"length"`
}

type Logging struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RuntimeIngest struct {
	Enabled     bool   `mapstructure:"enabled"`
	EndpointURL string `mapstructure:"endpoint_url"`
	Secret      string `mapstructure:"secret"`
	BearerToken string `mapstructure:"bearer_token"`
	IntervalSec int    `mapstructure:"interval_sec"`
}

type RelayConfig struct {
	Domain             string             `mapstructure:"domain"`
	ListenAddr         string             `mapstructure:"listen_addr"`
	TrustedProxies     []string           `mapstructure:"trusted_proxies"`
	ClientTokens       []RelayToken       `mapstructure:"client_tokens"`
	HostnameGeneration HostnameGeneration `mapstructure:"hostname_generation"`
	Logging            Logging            `mapstructure:"logging"`
	RuntimeIngest      RuntimeIngest      `mapstructure:"runtime_ingest"`
}

func LoadRelayConfig(path string) (*RelayConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetDefault("listen_addr", ":8080")
	v.SetDefault("hostname_generation.mode", "random")
	v.SetDefault("hostname_generation.length", 8)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("runtime_ingest.enabled", false)
	v.SetDefault("runtime_ingest.interval_sec", 30)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg RelayConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
