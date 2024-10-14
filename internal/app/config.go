package app

import (
	"os"
	"strings"

	"github.com/jeremywohl/flatten"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultFleetDBClientID      = "fleetscheduler-serverservice-api" // FleetDB still uses the ServerService Client ID
	defaultConditionOrcClientID = "fleetscheduler-condition-api"
)

// Configuration values are first grabbed from the config file. They must be in the right place in order to be grabbed.
// Values are then grabbed from the ENV variables, anything found will be used to override values in the config file.
// Example: Setting Configuration.Endpoints.FleetDB.URL
// In the config file (as yaml); endpoints.fleetdb.url: http://fleetdb:8000
// As a ENV variable; ENDPOINTS_FLEETDB_URL=http://fleetdb:8000
type Configuration struct {
	// FacilityCode limits this flipflop to events in a facility.
	FacilityCode string `mapstructure:"facility"`

	// LogLevel is the app verbose logging level.
	// one of - info, debug, trace
	LogLevel string `mapstructure:"log_level"`

	// Holds all endpoints
	Endpoints Endpoints `mapstructure:"endpoints"`
}

type Endpoints struct {
	// FleetDBConfig defines the fleetdb client configuration parameters
	ConditionOrc ConfigOIDC `mapstructure:"conditionorc"`

	// FleetDBConfig defines the fleetdb client configuration parameters
	FleetDB ConfigOIDC `mapstructure:"fleetdb"`
}

type ConfigOIDC struct {
	URL              string   `mapstructure:"url"`
	OidcIssuerURL    string   `mapstructure:"oidc_issuer_url"`
	OidcAudienceURL  string   `mapstructure:"oidc_audience_url"`
	OidcClientSecret string   `mapstructure:"oidc_client_secret"`
	OidcClientID     string   `mapstructure:"oidc_client_id"`
	OidcClientScopes []string `mapstructure:"oidc_client_scopes"`
	Authenticate     bool     `mapstructure:"authenticate"`
}

func LoadConfiguration(cfgFilePath string) (*Configuration, error) {
	v := viper.New()
	cfg := &Configuration{}

	err := cfg.envBindVars(v)
	if err != nil {
		return nil, err
	}

	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	err = readInFile(v, cfg, cfgFilePath)
	if err != nil {
		return nil, err
	}

	err = cfg.validate()
	return cfg, err
}

// Reads in the cfgFile when available and overrides from environment variables.
func readInFile(v *viper.Viper, cfg *Configuration, path string) error {
	if cfg == nil {
		return ErrConfig
	}

	if path != "" {
		fh, err := os.Open(path)
		if err != nil {
			return errors.Wrap(ErrConfig, err.Error())
		}
		defer fh.Close()

		if err = v.ReadConfig(fh); err != nil {
			return errors.Wrap(ErrConfig, "ReadConfig error:"+err.Error())
		}
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("config")
		err := v.ReadInConfig()
		if err != nil {
			return err
		}
	}

	err := v.Unmarshal(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *Configuration) validate() error {
	if cfg == nil {
		return ErrConfig
	}

	if cfg.FacilityCode == "" {
		return errors.Wrap(ErrConfig, "no facility code")
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	err := cfg.Endpoints.ConditionOrc.validate(defaultConditionOrcClientID)
	if err != nil {
		return err
	}

	err = cfg.Endpoints.FleetDB.validate(defaultFleetDBClientID)
	if err != nil {
		return err
	}

	return nil
}

// envBindVars binds environment variables to the struct
// without a configuration file being unmarshalled,
// this is a workaround for a viper bug,
//
// This can be replaced by the solution in https://github.com/spf13/viper/pull/1429
// once that PR is merged.
func (cfg *Configuration) envBindVars(v *viper.Viper) error {
	envKeysMap := map[string]interface{}{}
	if err := mapstructure.Decode(cfg, &envKeysMap); err != nil {
		return err
	}

	// Flatten nested conf map
	flat, err := flatten.Flatten(envKeysMap, "", flatten.DotStyle)
	if err != nil {
		return errors.Wrap(err, "Unable to flatten config")
	}

	for k := range flat {
		if err := v.BindEnv(k); err != nil {
			return errors.Wrap(ErrConfig, "env var bind error: "+err.Error())
		}
	}

	return nil
}

func (cfg *ConfigOIDC) validate(defaultClientID string) error {
	if cfg == nil {
		return errors.Wrapf(ErrInvalidConfig, "missing %s OIDC config", defaultClientID)
	}

	if cfg.OidcClientID == "" {
		cfg.OidcClientID = defaultClientID
	}
	err := errors.Wrap(ErrInvalidConfig, cfg.OidcClientID)

	if cfg.URL == "" {
		return errors.Wrap(err, "endpoint")
	}

	if cfg.Authenticate {
		if cfg.OidcIssuerURL == "" {
			return errors.Wrap(err, "oidc_issuer_url")
		}
		if cfg.OidcAudienceURL == "" {
			return errors.Wrap(err, "oidc_audience_url")
		}
		if len(cfg.OidcClientScopes) == 0 {
			return errors.Wrap(err, "oidc_client_scopes")
		}
		if cfg.OidcClientSecret == "" {
			return errors.Wrap(err, "oidc_client_secret")
		}
	}

	return nil
}
