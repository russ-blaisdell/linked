// Package config handles linked configuration and credential storage.
// Credentials are stored at ~/.openclaw/credentials/linkedin/<profile>/creds.json
// to be consistent with how OpenClaw manages other channel credentials.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/russ-blaisdell/linked/internal/models"
)

const (
	defaultProfile  = "default"
	credsDirBase    = ".openclaw/credentials/linkedin"
	credsFilename   = "creds.json"
)

// CredentialsPath returns the path to the credentials file for the given profile.
func CredentialsPath(profile string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	if profile == "" {
		profile = defaultProfile
	}
	return filepath.Join(home, credsDirBase, profile, credsFilename), nil
}

// LoadCredentials reads credentials for the given profile.
// If profile is empty, the default profile is used.
func LoadCredentials(profile string) (*models.Credentials, error) {
	path, err := CredentialsPath(profile)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credentials found — run 'linked auth setup' to configure")
		}
		return nil, fmt.Errorf("reading credentials: %w", err)
	}

	var creds models.Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	if creds.LiAt == "" {
		return nil, fmt.Errorf("credentials are incomplete — run 'linked auth setup' to reconfigure")
	}

	return &creds, nil
}

// SaveCredentials writes credentials for the given profile, creating directories as needed.
func SaveCredentials(profile string, creds *models.Credentials) error {
	path, err := CredentialsPath(profile)
	if err != nil {
		return err
	}

	creds.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating credentials directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding credentials: %w", err)
	}

	// Write with restricted permissions — credentials are sensitive.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	return nil
}

// DeleteCredentials removes stored credentials for the given profile.
func DeleteCredentials(profile string) error {
	path, err := CredentialsPath(profile)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	return nil
}

// ListProfiles returns all configured LinkedIn profiles.
func ListProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(home, credsDirBase)
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing profiles: %w", err)
	}
	var profiles []string
	for _, e := range entries {
		if e.IsDir() {
			profiles = append(profiles, e.Name())
		}
	}
	return profiles, nil
}
