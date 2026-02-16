package config

import "time"

const CurrentVersion = 1

type Config struct {
	Version       int                `json:"version"`
	ActiveProfile string             `json:"active_profile,omitempty"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Profile struct {
	APIKey       string    `json:"api_key,omitempty"`
	APISecret    string    `json:"api_secret,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	LastLoginAt  time.Time `json:"last_login_at,omitempty"`
}

func Default() Config {
	return Config{
		Version:  CurrentVersion,
		Profiles: make(map[string]Profile),
	}
}
