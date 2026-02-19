package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/config"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
)

func executeCLICommand(t *testing.T, configPath string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv(updater.DisableEnvVar, "1")

	cmd := newRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"--config", configPath}, args...))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func saveTestConfig(t *testing.T, configPath string, cfg config.Config) {
	t.Helper()

	if err := config.NewFileStore(configPath).Save(cfg); err != nil {
		t.Fatalf("save test config: %v", err)
	}
}

func loadTestConfig(t *testing.T, configPath string) config.Config {
	t.Helper()

	cfg, err := config.NewFileStore(configPath).Load()
	if err != nil {
		t.Fatalf("load test config: %v", err)
	}
	return cfg
}

func TestAuthLoginValidationRules(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = config.Profile{
		APIKey:    "test_key",
		APISecret: "test_secret",
	}
	saveTestConfig(t, configPath, cfg)

	tests := []struct {
		name     string
		args     []string
		errMatch string
	}{
		{
			name:     "requires exactly one mode",
			args:     []string{"auth", "login"},
			errMatch: "exactly one login mode is required",
		},
		{
			name:     "rejects callback with request-token",
			args:     []string{"auth", "login", "--callback", "--request-token", "abc"},
			errMatch: "--request-token cannot be used with --callback",
		},
		{
			name:     "rejects callback-port without callback",
			args:     []string{"auth", "login", "--request-token", "abc", "--callback-port", "9999"},
			errMatch: "--callback-port can only be used with --callback",
		},
		{
			name:     "rejects callback-port lower than range",
			args:     []string{"auth", "login", "--callback", "--callback-port", "0"},
			errMatch: "--callback-port must be between 1 and 65535",
		},
		{
			name:     "rejects callback-port higher than range",
			args:     []string{"auth", "login", "--callback", "--callback-port", "65536"},
			errMatch: "--callback-port must be between 1 and 65535",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _, err := executeCLICommand(t, configPath, tc.args...)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.errMatch) {
				t.Fatalf("expected error to contain %q, got %q", tc.errMatch, err.Error())
			}
			if strings.Contains(stdout, "Login URL:") {
				t.Fatalf("expected validation failure before login URL is printed, got output: %q", stdout)
			}
		})
	}
}

func TestAuthLoginRequestTokenURLModePassesValidation(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = config.Profile{
		APIKey: "test_key",
	}
	saveTestConfig(t, configPath, cfg)

	_, _, err := executeCLICommand(
		t,
		configPath,
		"auth",
		"login",
		"--request-token",
		"https://kite.zerodha.com/connect/login?request_token=abc123&status=success",
	)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "profile missing api_secret") {
		t.Fatalf("expected missing api_secret error, got %q", err.Error())
	}
}

func TestValidateAuthLoginFlagsExtractsRequestTokenFromURL(t *testing.T) {
	token, err := validateAuthLoginFlags(
		false,
		"https://kite.zerodha.com/connect/login?request_token=abc123&status=success",
		8787,
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc123" {
		t.Fatalf("expected extracted token %q, got %q", "abc123", token)
	}
}
