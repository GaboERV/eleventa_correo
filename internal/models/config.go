package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the local client configuration for the email-based reporter.
type Config struct {
	BranchName  string `json:"branch_name"`
	SmtpUser    string `json:"smtp_user"`
	SmtpPass    string `json:"smtp_pass"`
	TargetEmail string `json:"target_email"`
	AutoDay     string `json:"auto_day"`  // e.g. "Monday"
	AutoTime    string `json:"auto_time"` // e.g. "08:00"
}

// LoadConfig reads config.json from the given directory.
func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config.json: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config.json: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes config.json to the given directory.
func SaveConfig(dir string, cfg *Config) error {
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing config.json: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config.json: %w", err)
	}
	return nil
}

// LastRun tracks the last successful report window to ensure idempotency.
type LastRun struct {
	LastSuccessfulReportEnd string `json:"last_successful_report_end"`
	LastReportedTurnoID     int    `json:"last_reported_turno_id"`
}

// LoadLastRun reads last_run.json from the given directory.
func LoadLastRun(dir string) (*LastRun, error) {
	path := filepath.Join(dir, "last_run.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LastRun{}, nil // First run, no previous state
		}
		return nil, fmt.Errorf("error reading last_run.json: %w", err)
	}
	var lr LastRun
	if err := json.Unmarshal(data, &lr); err != nil {
		return nil, fmt.Errorf("error parsing last_run.json: %w", err)
	}
	return &lr, nil
}

// SaveLastRun writes last_run.json to the given directory.
func SaveLastRun(dir string, lr *LastRun) error {
	path := filepath.Join(dir, "last_run.json")
	data, err := json.MarshalIndent(lr, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing last_run.json: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing last_run.json: %w", err)
	}
	return nil
}
