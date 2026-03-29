package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading non-existent file
	cfg, err := Load("/non/existent/path")
	if err != nil {
		t.Errorf("Load non-existent file: expected no error, got %v", err)
	}
	if cfg == nil {
		t.Errorf("Load non-existent file: expected non-nil Config, got nil")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("Load non-existent file: expected empty profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("Load non-existent file: expected empty ActiveProfile, got %s", cfg.ActiveProfile)
	}

	// Test loading valid JSON file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create valid config
	validConfig := `{
		"active_profile": "test",
		"profiles": [
			{
				"name": "test",
				"region": "us-east-1",
				"access_key_id": "testkey",
				"secret_access_key": "testsecret",
				"bucket": "testbucket",
				"endpoint_url": "http://localhost:9000"
			}
		]
	}`
	
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	
	cfg, err = Load(configPath)
	if err != nil {
		t.Errorf("Load valid config: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Errorf("Load valid config: expected non-nil Config, got nil")
	}
	if cfg.ActiveProfile != "test" {
		t.Errorf("Load valid config: expected ActiveProfile 'test', got %s", cfg.ActiveProfile)
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("Load valid config: expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Name != "test" {
		t.Errorf("Load valid config: expected profile name 'test', got %s", cfg.Profiles[0].Name)
	}

	// Test loading invalid JSON file
	invalidConfig := `{
		"active_profile": "test",
		"profiles": [
			{
				"name": "test",
				"region": "us-east-1"
			]
		}
	}`
	
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid test config: %v", err)
	}
	
	_, err = Load(configPath)
	if err == nil {
		t.Errorf("Load invalid config: expected error, got nil")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Test saving valid config
	cfg := &Config{
		ActiveProfile: "test",
		Profiles: []Profile{
			{
				Name:            "test",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
				Bucket:          "testbucket",
				EndpointURL:     "http://localhost:9000",
			},
		},
	}
	
	if err := Save(cfg, configPath); err != nil {
		t.Errorf("Save valid config: unexpected error: %v", err)
	}
	
	// Verify saved file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Read saved config: unexpected error: %v", err)
	}
	
	var loadedCfg Config
	if err := json.Unmarshal(data, &loadedCfg); err != nil {
		t.Errorf("Unmarshal saved config: unexpected error: %v", err)
	}
	
	if loadedCfg.ActiveProfile != "test" {
		t.Errorf("Saved config: expected ActiveProfile 'test', got %s", loadedCfg.ActiveProfile)
	}
	if len(loadedCfg.Profiles) != 1 {
		t.Errorf("Saved config: expected 1 profile, got %d", len(loadedCfg.Profiles))
	}
	if loadedCfg.Profiles[0].Name != "test" {
		t.Errorf("Saved config: expected profile name 'test', got %s", loadedCfg.Profiles[0].Name)
	}
	
	// Test saving with non-existent directory (should create it)
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.json")
	if err := Save(cfg, nestedPath); err != nil {
		t.Errorf("Save config to nested directory: unexpected error: %v", err)
	}
	
	// Verify nested file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Errorf("Saved config to nested directory: file does not exist")
	}
}

func TestGetActiveProfile(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "active",
		Profiles: []Profile{
			{
				Name: "inactive",
			},
			{
				Name: "active",
			},
			{
				Name: "another",
			},
		},
	}
	
	profile := GetActiveProfile(cfg)
	if profile == nil {
		t.Errorf("GetActiveProfile: expected non-nil profile, got nil")
		return
	}
	if profile.Name != "active" {
		t.Errorf("GetActiveProfile: expected profile name 'active', got %s", profile.Name)
	}
	
	// Test with non-active profile
	cfg.ActiveProfile = "inactive"
	profile = GetActiveProfile(cfg)
	if profile == nil {
		t.Errorf("GetActiveProfile: expected non-nil profile for inactive, got nil")
		return
	}
	if profile.Name != "inactive" {
		t.Errorf("GetActiveProfile: expected profile name 'inactive', got %s", profile.Name)
	}
	
	// Test with non-existent profile
	cfg.ActiveProfile = "nonexistent"
	profile = GetActiveProfile(cfg)
	if profile != nil {
		t.Errorf("GetActiveProfile: expected nil profile for nonexistent, got %v", profile)
	}
	
	// Test with empty config
	cfg.ActiveProfile = ""
	cfg.Profiles = []Profile{}
	profile = GetActiveProfile(cfg)
	if profile != nil {
		t.Errorf("GetActiveProfile: expected nil profile for empty config, got %v", profile)
	}
}

func TestFindProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{
				Name: "first",
			},
			{
				Name: "second",
			},
			{
				Name: "third",
			},
		},
	}
	
	// Test finding existing profile
	profile := FindProfile(cfg, "second")
	if profile == nil {
		t.Errorf("FindProfile: expected non-nil profile for 'second', got nil")
		return
	}
	if profile.Name != "second" {
		t.Errorf("FindProfile: expected profile name 'second', got %s", profile.Name)
	}
	
	// Test finding first profile
	profile = FindProfile(cfg, "first")
	if profile == nil {
		t.Errorf("FindProfile: expected non-nil profile for 'first', got nil")
		return
	}
	if profile.Name != "first" {
		t.Errorf("FindProfile: expected profile name 'first', got %s", profile.Name)
	}
	
	// Test finding last profile
	profile = FindProfile(cfg, "third")
	if profile == nil {
		t.Errorf("FindProfile: expected non-nil profile for 'third', got nil")
		return
	}
	if profile.Name != "third" {
		t.Errorf("FindProfile: expected profile name 'third', got %s", profile.Name)
	}
	
	// Test finding non-existent profile
	profile = FindProfile(cfg, "nonexistent")
	if profile != nil {
		t.Errorf("FindProfile: expected nil profile for 'nonexistent', got %v", profile)
	}
	
	// Test with empty profiles
	cfg.Profiles = []Profile{}
	profile = FindProfile(cfg, "any")
	if profile != nil {
		t.Errorf("FindProfile: expected nil profile for empty list, got %v", profile)
	}
}