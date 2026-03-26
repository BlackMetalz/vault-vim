package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	VaultAddr  string
	VaultToken string
}

func Load() (*Config, error) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:8200"
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		token = readTokenFile()
	}

	if token == "" {
		return nil, fmt.Errorf("no vault token found: set VAULT_TOKEN or run 'vault login'")
	}

	return &Config{
		VaultAddr:  addr,
		VaultToken: token,
	}, nil
}

func readTokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".vault-token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
