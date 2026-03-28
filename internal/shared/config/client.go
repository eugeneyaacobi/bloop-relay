package config

import "github.com/spf13/viper"

type ReconnectConfig struct {
	InitialDelayMs int `mapstructure:"initial_delay_ms"`
	MaxDelayMs     int `mapstructure:"max_delay_ms"`
}

type BasicAuthConfig struct {
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	PasswordEnv string `mapstructure:"password_env"`
}

type TunnelConfig struct {
	Name      string          `mapstructure:"name"`
	Hostname  string          `mapstructure:"hostname"`
	LocalAddr string          `mapstructure:"local_addr"`
	Access    string          `mapstructure:"access"`
	Token     string          `mapstructure:"token"`
	TokenEnv  string          `mapstructure:"token_env"`
	BasicAuth BasicAuthConfig `mapstructure:"basic_auth"`
}

type ClientConfig struct {
	RelayURL          string          `mapstructure:"relay_url"`
	AuthToken         string          `mapstructure:"auth_token"`
	AuthTokenEnv      string          `mapstructure:"auth_token_env"`
	ControlPlaneURL   string          `mapstructure:"control_plane_url"`
	EnrollmentToken   string          `mapstructure:"enrollment_token"`
	EnrollmentTokenEnv string         `mapstructure:"enrollment_token_env"`
	Reconnect         ReconnectConfig `mapstructure:"reconnect"`
	Tunnels           []TunnelConfig  `mapstructure:"tunnels"`
	Logging           Logging         `mapstructure:"logging"`
}

func LoadClientConfig(path string) (*ClientConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetDefault("reconnect.initial_delay_ms", 1000)
	v.SetDefault("reconnect.max_delay_ms", 30000)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg ClientConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
