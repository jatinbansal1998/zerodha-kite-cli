package cli

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/config"
)

func TestConfigProfileAddUpsertOverwritesCredentials(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"add",
		"default",
		"--api-key",
		"old_key",
		"--api-secret",
		"old_secret",
		"--set-active",
	); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"add",
		"default",
		"--api-key",
		"new_key",
		"--api-secret",
		"new_secret",
	); err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	cfg := loadTestConfig(t, configPath)
	profile, ok := cfg.Profiles["default"]
	if !ok {
		t.Fatalf("expected profile %q to exist", "default")
	}
	if profile.APIKey != "new_key" {
		t.Fatalf("expected API key %q, got %q", "new_key", profile.APIKey)
	}
	if profile.APISecret != "new_secret" {
		t.Fatalf("expected API secret %q, got %q", "new_secret", profile.APISecret)
	}
	if cfg.ActiveProfile != "default" {
		t.Fatalf("expected active profile %q, got %q", "default", cfg.ActiveProfile)
	}
}

func TestConfigProfileSetAPIKeyUpdatesOnlyKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	loginTime := time.Date(2025, 2, 10, 9, 30, 0, 0, time.UTC)

	cfg := config.Default()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = config.Profile{
		APIKey:       "old_key",
		APISecret:    "secret",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		LastLoginAt:  loginTime,
	}
	saveTestConfig(t, configPath, cfg)

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"set-api-key",
		"default",
		"--api-key",
		"new_key",
	); err != nil {
		t.Fatalf("set-api-key failed: %v", err)
	}

	updated := loadTestConfig(t, configPath)
	profile := updated.Profiles["default"]
	if profile.APIKey != "new_key" {
		t.Fatalf("expected API key %q, got %q", "new_key", profile.APIKey)
	}
	if profile.APISecret != "secret" {
		t.Fatalf("expected API secret %q, got %q", "secret", profile.APISecret)
	}
	if profile.AccessToken != "access_token" {
		t.Fatalf("expected access token %q, got %q", "access_token", profile.AccessToken)
	}
	if profile.RefreshToken != "refresh_token" {
		t.Fatalf("expected refresh token %q, got %q", "refresh_token", profile.RefreshToken)
	}
	if !profile.LastLoginAt.Equal(loginTime) {
		t.Fatalf("expected last_login_at %s, got %s", loginTime.Format(time.RFC3339), profile.LastLoginAt.Format(time.RFC3339))
	}
}

func TestConfigProfileSetAPISecretUpdatesOnlySecret(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	loginTime := time.Date(2025, 2, 10, 9, 30, 0, 0, time.UTC)

	cfg := config.Default()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = config.Profile{
		APIKey:       "key",
		APISecret:    "old_secret",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		LastLoginAt:  loginTime,
	}
	saveTestConfig(t, configPath, cfg)

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"set-api-secret",
		"default",
		"--api-secret",
		"new_secret",
	); err != nil {
		t.Fatalf("set-api-secret failed: %v", err)
	}

	updated := loadTestConfig(t, configPath)
	profile := updated.Profiles["default"]
	if profile.APISecret != "new_secret" {
		t.Fatalf("expected API secret %q, got %q", "new_secret", profile.APISecret)
	}
	if profile.APIKey != "key" {
		t.Fatalf("expected API key %q, got %q", "key", profile.APIKey)
	}
	if profile.AccessToken != "access_token" {
		t.Fatalf("expected access token %q, got %q", "access_token", profile.AccessToken)
	}
	if profile.RefreshToken != "refresh_token" {
		t.Fatalf("expected refresh token %q, got %q", "refresh_token", profile.RefreshToken)
	}
	if !profile.LastLoginAt.Equal(loginTime) {
		t.Fatalf("expected last_login_at %s, got %s", loginTime.Format(time.RFC3339), profile.LastLoginAt.Format(time.RFC3339))
	}
}

func TestConfigProfileImplicitActiveProfileBehavior(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"add",
		"alpha",
		"--api-key",
		"alpha_key",
		"--api-secret",
		"alpha_secret",
	); err != nil {
		t.Fatalf("add alpha failed: %v", err)
	}

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"add",
		"beta",
		"--api-key",
		"beta_key",
		"--api-secret",
		"beta_secret",
	); err != nil {
		t.Fatalf("add beta failed: %v", err)
	}

	cfg := loadTestConfig(t, configPath)
	if cfg.ActiveProfile != "alpha" {
		t.Fatalf("expected active profile to remain %q after second add, got %q", "alpha", cfg.ActiveProfile)
	}

	if _, _, err := executeCLICommand(
		t,
		configPath,
		"config",
		"profile",
		"remove",
		"alpha",
	); err != nil {
		t.Fatalf("remove alpha failed: %v", err)
	}

	updated := loadTestConfig(t, configPath)
	if updated.ActiveProfile != "beta" {
		t.Fatalf("expected active profile to auto-switch to %q, got %q", "beta", updated.ActiveProfile)
	}
}
