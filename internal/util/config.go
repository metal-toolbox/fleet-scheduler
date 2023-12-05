package util

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const config_env_variable_name = "FLEETSCHEDULERCONFIG"

type Configuration struct {
	// LogLevel is the app verbose logging level.
	// one of - info, debug, trace
	LogLevel string `mapstructure:"log_level"`

	// CSV file path when StoreKind is set to csv.
	CsvFile string `mapstructure:"csv_file"`

	// FacilityCode limits this alloy to events in a facility.
	FacilityCode string `mapstructure:"facility_code"`

	// ServerserviceConfig defines the serverservice client configuration parameters
	SsCfg *ServerserviceConfig `mapstructure:"serverservice"`
}

type ServerserviceConfig struct {
	EndpointURL          *url.URL
	FacilityCode         string   `mapstructure:"facility_code"`
	Endpoint             string   `mapstructure:"endpoint"`
	OidcIssuerEndpoint   string   `mapstructure:"oidc_issuer_endpoint"`
	OidcAudienceEndpoint string   `mapstructure:"oidc_audience_endpoint"`
	OidcClientSecret     string   `mapstructure:"oidc_client_secret"`
	OidcClientID         string   `mapstructure:"oidc_client_id"`
	OidcClientScopes     []string `mapstructure:"oidc_client_scopes"`
	DisableOAuth         bool     `mapstructure:"disable_oauth"`
}

func LoadConfig(path string) (*Configuration, error) {
	cfg := &Configuration{}
	viper.AutomaticEnv()
	h, err := openConfig(path)
	if err != nil {
		return cfg, err
	}

	pathName := h.Name()
	viper.SetConfigFile(pathName)

	err = viper.ReadConfig(h)
	if err != nil {
		return cfg, errors.Wrap(err, pathName)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func openConfig(path string) (*os.File, error) {
	if path != "" {
		return os.Open(path)
	}
	path = viper.GetString(config_env_variable_name)
	if path != "" {
		return os.Open(path)
	}

	path = filepath.Join(xdg.Home, ".mctl.yml")
	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	path, err = xdg.ConfigFile("mctl/config.yaml")
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}