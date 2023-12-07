package app

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	config_env_variable_name = "FLEET_SCHEDULER_CONFIG"

	LogLevelInfo = "info"
	LogLevelDebug = "debug"
	LogLevelTrace = "trace"
)

type Configuration struct {
	// LogLevel is the app verbose logging level.
	// one of - info, debug, trace
	LogLevel string `mapstructure:"log_level"`

	// CSV file path when StoreKind is set to csv.
	CsvFile string `mapstructure:"csv_file"`

	// FacilityCode limits this alloy to events in a facility.
	FacilityCode string `mapstructure:"facility_code"`

	// FleetDBConfig defines the fleetdb (serverservice) client configuration parameters
	FdbCfg *FleetDBConfig `mapstructure:"fleetdb"`
}

type FleetDBConfig struct {
	FacilityCode         string   `mapstructure:"facility_code"`
	Endpoint             string   `mapstructure:"endpoint"`
	OidcIssuerEndpoint   string   `mapstructure:"oidc_issuer_endpoint"`
	OidcAudienceEndpoint string   `mapstructure:"oidc_audience_endpoint"`
	OidcClientSecret     string   `mapstructure:"oidc_client_secret"`
	OidcClientID         string   `mapstructure:"oidc_client_id"`
	OidcClientScopes     []string `mapstructure:"oidc_client_scopes"`
	DisableOAuth         bool     `mapstructure:"disable_oauth"`
}

func loadConfig(path string) (*Configuration, error) {
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

	err = validateClientParams(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateClientParams(cfg *Configuration) error {
	errCfgInvalid := errors.New("Configuration is invalid")
	errFleetDbInvalid := errors.Wrap(errCfgInvalid, "FleetDbCfg")

	if cfg.LogLevel == "" {
		cfg.LogLevel = LogLevelInfo
	} else {
		if cfg.LogLevel != LogLevelInfo &&
			cfg.LogLevel != LogLevelDebug &&
			cfg.LogLevel != LogLevelTrace {
				return errors.Wrap(errCfgInvalid, "LogLevel")
			}
	}

	// FleetDB (serverservice) Configuration
	if cfg.FdbCfg == nil {
		return errFleetDbInvalid
	}
	if cfg.FdbCfg.FacilityCode == "" {
		return errors.Wrap(errFleetDbInvalid, "Facility Code")
	}
	if cfg.FdbCfg.Endpoint == "" {
		return errors.Wrap(errFleetDbInvalid, "Endpoint")
	}

	// OAUTH
	if !cfg.FdbCfg.DisableOAuth {
		errCfgOIDCConfig := errors.New("OIDC is invalide")

		if cfg.FdbCfg.OidcIssuerEndpoint == "" {
			return errors.Wrap(errCfgOIDCConfig, "Issuer Endpoint")
		}
		if cfg.FdbCfg.OidcAudienceEndpoint == "" {
			return errors.Wrap(errCfgOIDCConfig, "Audience Endpoint")
		}
		if cfg.FdbCfg.OidcClientSecret == "" {
			return errors.Wrap(errCfgOIDCConfig, "Client Secret")
		}
		if cfg.FdbCfg.OidcClientID == "" {
			return errors.Wrap(errCfgOIDCConfig, "Client ID")
		}
		if len(cfg.FdbCfg.OidcClientScopes) == 0 {
			return errors.Wrap(errCfgOIDCConfig, "Client Scopes")
		}
	}

	return nil
}

func openConfig(path string) (*os.File, error) {
	if path != "" {
		return os.Open(path)
	}
	path = viper.GetString(config_env_variable_name)
	if path != "" {
		return os.Open(path)
	}

	path = filepath.Join(xdg.Home, ".config.yaml")
	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return os.Open(path)
}