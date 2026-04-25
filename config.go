package main

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultBindAddr = ":8080"
	defaultUserName = "local/loxone-bridge"
	defaultDataDir  = ".data"
)

type config struct {
	BindAddr           string
	HomeWizardHost     string
	HomeWizardUserName string
	DataDir            string
	Token              string
	InsecureSkipVerify bool
}

func loadConfig() (config, error) {
	cfg := config{
		BindAddr:           getenvDefault("BIND_ADDR", defaultBindAddr),
		HomeWizardHost:     strings.TrimSpace(os.Getenv("HOMEWIZARD_HOST")),
		HomeWizardUserName: getenvDefault("HOMEWIZARD_USERNAME", defaultUserName),
		DataDir:            getenvDefault("DATA_DIR", defaultDataDir),
		Token:              strings.TrimSpace(os.Getenv("HOMEWIZARD_TOKEN")),
		InsecureSkipVerify: getenvBoolDefault("HOMEWIZARD_INSECURE_SKIP_VERIFY", true),
	}

	if cfg.HomeWizardHost == "" {
		return config{}, errors.New("HOMEWIZARD_HOST is required")
	}

	cfg.HomeWizardHost = strings.TrimRight(cfg.HomeWizardHost, "/")
	if !strings.Contains(cfg.HomeWizardHost, "://") {
		cfg.HomeWizardHost = "http://" + cfg.HomeWizardHost
	}

	if !strings.HasPrefix(cfg.HomeWizardUserName, "local/") {
		return config{}, errors.New("HOMEWIZARD_USERNAME must start with local/")
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvBoolDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
