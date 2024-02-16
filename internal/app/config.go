package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	appName string = "fleet-scheduler"

	defaultFleetDBClientID      = "fleetscheduler-serverservice-api"
	defaultConditionOrcClientID = "fleetscheduler-condition-api"

	defaultConcurrencyCount = 4

	configEnvVariableName = "FLEET_SCHEDULER_CONFIG"
)

type Configuration struct {
	// LogLevel is the app verbose logging level.
	// one of - info, debug, trace
	LogLevel string `mapstructure:"log_level"`

	// FacilityCode limits this fleet scheduler to events in a facility.
	FacilityCode string `mapstructure:"facility_code"`

	// Max threads allowed for communicating with other resources.
	Concurrency int `mapstructure:"concurrency"`

	// Defines the fleetdb (serverservice) client configuration parameters
	FdbCfg *ConfigOIDC `mapstructure:"fleetdb_api"`
	// Defines the condition orchestrator client configuration parameters
	CoCfg *ConfigOIDC `mapstructure:"conditionorc_api"`
}

type ConfigOIDC struct {
	// Skips OAuth setup if true
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

func LoadConfig(path string) (*Configuration, error) {
	cfg := &Configuration{}
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

func loadEnvOverrides(cfg *Configuration, v *viper.Viper) error {
	if !cfg.FdbCfg.DisableOAuth {
		cfg.FdbCfg.ClientSecret = v.GetString("fleetdb.oidc.client.secret")
		if cfg.FdbCfg.ClientSecret == "" {
			return errors.New("FLEET_SCHEDULER_FLEETDB_OIDC_CLIENT_SECRET was empty")
		}
	}

	if !cfg.CoCfg.DisableOAuth {
		cfg.CoCfg.ClientSecret = v.GetString("conditionorc.oidc.client.secret")
		if cfg.FdbCfg.ClientSecret == "" {
			return errors.New("FLEET_SCHEDULER_CONDITIONORC_OIDC_CLIENT_SECRET was empty")
		}
	}

	return nil
}

func validateClientParams(cfg *Configuration) error {
	if cfg.LogLevel == "" {
		cfg.LogLevel = "debug"
	}

	if cfg.Concurrency <= 0 {
		cfg.Concurrency = defaultConcurrencyCount
	}

	// FleetDB (serverservice) Configuration
	if cfg.FdbCfg == nil {
		return errors.Wrap(ErrInvalidConfig, "fleetdb_api entry doesnt exist")
	}
	if cfg.CoCfg == nil {
		return errors.Wrap(ErrInvalidConfig, "conditionorc_api entry doesnt exist")
	}
	if cfg.FacilityCode == "" {
		return errors.Wrap(ErrInvalidConfig, "Facility Code")
	}

	v := viper.New()
	v.SetEnvPrefix(appName)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	err := loadEnvOverrides(cfg, v)
	if err != nil {
		return err
	}

	err = validateOIDCConfig(cfg.FdbCfg, defaultFleetDBClientID)
	if err != nil {
		return err
	}
	err = validateOIDCConfig(cfg.CoCfg, defaultConditionOrcClientID)
	if err != nil {
		return err
	}

	return nil
}

func validateOIDCConfig(cfg *ConfigOIDC, defaultClientID string) error {
	if cfg.ClientID == "" {
		cfg.ClientID = defaultClientID
	}
	err := errors.Wrap(ErrInvalidConfig, cfg.ClientID)

	if cfg.Endpoint == "" {
		return errors.Wrap(err, "endpoint")
	}

	if !cfg.DisableOAuth {
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
