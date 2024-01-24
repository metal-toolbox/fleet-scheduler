package app

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	configEnvVariableName = "FLEET_SCHEDULER_CONFIG"

	LogLevelInfo  = "info"
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
	FdbCfg *ConfigOIDC `mapstructure:"fleetdb_api"`
	CoCfg  *ConfigOIDC `mapstructure:"conditionorc_api"`
}

type ConfigOIDC struct {
	// Disable skips OAuth setup
	DisableOAuth bool `mapstructure:"disable_oauth"`

	// ServerService OAuth2 parameters
	Endpoint         string   `mapstructure:"endpoint"`
	ClientID         string   `mapstructure:"oidc_client_id"`
	IssuerEndpoint   string   `mapstructure:"oidc_issuer_endpoint"`
	AudienceEndpoint string   `mapstructure:"oidc_audience_endpoint"`
	ClientScopes     []string `mapstructure:"oidc_scopes"`
	PkceCallbackURL  string   `mapstructure:"oidc_pkce_callback_url"`
	ClientSecret     string   `mapstructure:"oidc_client_secret"`
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

	if cfg.LogLevel == "" {
		cfg.LogLevel = LogLevelInfo
	} else if cfg.LogLevel != LogLevelInfo &&
		cfg.LogLevel != LogLevelDebug &&
		cfg.LogLevel != LogLevelTrace {
		return errors.Wrap(errCfgInvalid, "LogLevel")
	}

	// FleetDB (serverservice) Configuration
	if cfg.FdbCfg == nil {
		return errors.Wrap(errCfgInvalid, "fleetdb_api entry doesnt exist")
	}
	if cfg.CoCfg == nil {
		return errors.Wrap(errCfgInvalid, "conditionorc_api entry doesnt exist")
	}
	if cfg.FacilityCode == "" {
		return errors.Wrap(errCfgInvalid, "Facility Code")
	}
	err := validateOIDCConfig(cfg.FdbCfg, errors.Wrap(errCfgInvalid, "fleetdb_api is invalid"))
	if err != nil {
		return err
	}
	err = validateOIDCConfig(cfg.CoCfg, errors.Wrap(errCfgInvalid, "conditionorc_api is invalid"))
	if err != nil {
		return err
	}

	return nil
}

func validateOIDCConfig(cfg *ConfigOIDC, err error) error {
	if cfg.Endpoint == "" {
		return errors.Wrap(err, "endpoint")
	}

	if !cfg.DisableOAuth {
		if cfg.ClientID == "" {
			return errors.Wrap(err, "oidc_client_id")
		}
		if cfg.IssuerEndpoint == "" {
			return errors.Wrap(err, "oidc_issuer_endpoint")
		}
		if cfg.AudienceEndpoint == "" {
			return errors.Wrap(err, "oidc_audience_endpoint")
		}
		if len(cfg.ClientScopes) == 0 {
			return errors.Wrap(err, "oidc_client_scopes")
		}
		if cfg.ClientSecret == "" {
			return errors.Wrap(err, "oidc_client_secret")
		}
		if cfg.PkceCallbackURL == "" {
			return errors.Wrap(err, "oidc_pkce_callback_url")
		}
	}

	return nil
}

func openConfig(path string) (*os.File, error) {
	if path != "" {
		return os.Open(path)
	}
	path = viper.GetString(configEnvVariableName)
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
