package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_addProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create initial config
	if err := CreateDefaultConfig(configPath); err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}
	
	cli := NewCLI(configPath)
	
	// Test adding profile with all required flags
	args := []string{
		"--name", "test-profile",
		"--region", "us-east-1",
		"--access-key", "testkey123",
		"--secret", "testsecret456",
		"--endpoint", "http://localhost:9000",
		"--bucket", "test-bucket",
	}
	
	if err := cli.Run(append([]string{"add"}, args...)); err != nil {
		t.Errorf("CLI addProfile: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after add: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("CLI addProfile: expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "test-profile" {
		t.Errorf("CLI addProfile: expected ActiveProfile 'test-profile', got %s", cfg.ActiveProfile)
	}
	if cfg.Profiles[0].Name != "test-profile" {
		t.Errorf("CLI addProfile: expected profile name 'test-profile', got %s", cfg.Profiles[0].Name)
	}
	if cfg.Profiles[0].Region != "us-east-1" {
		t.Errorf("CLI addProfile: expected region 'us-east-1', got %s", cfg.Profiles[0].Region)
	}
	if cfg.Profiles[0].AccessKeyID != "testkey123" {
		t.Errorf("CLI addProfile: expected access key 'testkey123', got %s", cfg.Profiles[0].AccessKeyID)
	}
	if cfg.Profiles[0].SecretAccessKey != "testsecret456" {
		t.Errorf("CLI addProfile: expected secret key 'testsecret456', got %s", cfg.Profiles[0].SecretAccessKey)
	}
	if cfg.Profiles[0].EndpointURL != "http://localhost:9000" {
		t.Errorf("CLI addProfile: expected endpoint 'http://localhost:9000', got %s", cfg.Profiles[0].EndpointURL)
	}
	if cfg.Profiles[0].Bucket != "test-bucket" {
		t.Errorf("CLI addProfile: expected bucket 'test-bucket', got %s", cfg.Profiles[0].Bucket)
	}
	
	// Test adding profile missing required flags
	args = []string{
		"--name", "test-profile-2",
		"--region", "us-west-2",
		"--access-key", "testkey789",
		// missing --secret
	}
	
	if err := cli.Run(append([]string{"add"}, args...)); err == nil {
		t.Errorf("CLI addProfile missing secret: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "required flags") {
			t.Errorf("CLI addProfile missing secret: unexpected error: %v", err)
		}
	}
	
	// Test adding duplicate profile
	args = []string{
		"--name", "test-profile",
		"--region", "eu-central-1",
		"--access-key", "anotherkey",
		"--secret", "anothersecret",
	}
	
	if err := cli.Run(append([]string{"add"}, args...)); err == nil {
		t.Errorf("CLI addProfile duplicate: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile already exists") {
			t.Errorf("CLI addProfile duplicate: unexpected error: %v", err)
		}
	}
}

func TestCLI_listProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create config with profiles
	cfg := &Config{
		ActiveProfile: "active-profile",
		Profiles: []Profile{
			{
				Name:            "inactive-profile",
				Region:          "us-east-1",
				AccessKeyID:     "inactivekey",
				SecretAccessKey: "inactiversecret",
			},
			{
				Name:            "active-profile",
				Region:          "us-west-2",
				AccessKeyID:     "activekey",
				SecretAccessKey: "activesecret",
				EndpointURL:     "http://custom-endpoint:9000",
				Bucket:          "active-bucket",
			},
		},
	}
	
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	cli := NewCLI(configPath)
	
	// Capture output
	output := captureOutput(t, func() error {
		return cli.Run([]string{"list"})
	})
	
	if !strings.Contains(output, "Profiles:") {
		t.Errorf("CLI listProfiles: expected 'Profiles:' in output, got %q", output)
	}
	if !strings.Contains(output, "inactive-profile") {
		t.Errorf("CLI listProfiles: expected 'inactive-profile' in output, got %q", output)
	}
	if !strings.Contains(output, "active-profile (active)") {
		t.Errorf("CLI listProfiles: expected 'active-profile (active)' in output, got %q", output)
	}
	if !strings.Contains(output, "Endpoint: http://custom-endpoint:9000") {
		t.Errorf("CLI listProfiles: expected endpoint in output, got %q", output)
	}
	if !strings.Contains(output, "Region: us-east-1, Bucket: ") {
		t.Errorf("CLI listProfiles: expected region/bucket for inactive profile, got %q", output)
	}
	if !strings.Contains(output, "Region: us-west-2, Bucket: active-bucket") {
		t.Errorf("CLI listProfiles: expected region/bucket for active profile, got %q", output)
	}
	
	// Test empty profiles list
	cfg.Profiles = []Profile{}
	cfg.ActiveProfile = ""
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save empty config: %v", err)
	}
	
	output = captureOutput(t, func() error {
		return cli.Run([]string{"list"})
	})
	
	if !strings.Contains(output, "No profiles configured") {
		t.Errorf("CLI listProfiles empty: expected 'No profiles configured' in output, got %q", output)
	}
}

func TestCLI_switchProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create config with profiles
	cfg := &Config{
		ActiveProfile: "first-profile",
		Profiles: []Profile{
			{
				Name: "first-profile",
			},
			{
				Name: "second-profile",
			},
			{
				Name: "third-profile",
			},
		},
	}
	
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	cli := NewCLI(configPath)
	
	// Test switching to second profile
	if err := cli.Run([]string{"switch", "--name", "second-profile"}); err != nil {
		t.Errorf("CLI switchProfile to second: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after switch: %v", err)
	}
	if cfg.ActiveProfile != "second-profile" {
		t.Errorf("CLI switchProfile: expected ActiveProfile 'second-profile', got %s", cfg.ActiveProfile)
	}
	
	// Test switching to third profile
	if err := cli.Run([]string{"switch", "--name", "third-profile"}); err != nil {
		t.Errorf("CLI switchProfile to third: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err = Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after switch: %v", err)
	}
	if cfg.ActiveProfile != "third-profile" {
		t.Errorf("CLI switchProfile: expected ActiveProfile 'third-profile', got %s", cfg.ActiveProfile)
	}
	
	// Test switching with missing name flag
	if err := cli.Run([]string{"switch"}); err == nil {
		t.Errorf("CLI switchProfile missing name: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile name is required") {
			t.Errorf("CLI switchProfile missing name: unexpected error: %v", err)
		}
	}
	
	// Test switching to non-existent profile
	if err := cli.Run([]string{"switch", "--name", "nonexistent"}); err == nil {
		t.Errorf("CLI switchProfile nonexistent: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("CLI switchProfile nonexistent: unexpected error: %v", err)
		}
	}
}

func TestCLI_removeProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create config with profiles
	cfg := &Config{
		ActiveProfile: "middle-profile",
		Profiles: []Profile{
			{
				Name: "first-profile",
			},
			{
				Name: "middle-profile",
			},
			{
				Name: "last-profile",
			},
		},
	}
	
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	cli := NewCLI(configPath)
	
	// Test removing middle profile (active)
	if err := cli.Run([]string{"remove", "--name", "middle-profile"}); err != nil {
		t.Errorf("CLI removeProfile middle: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after remove: %v", err)
	}
	if len(cfg.Profiles) != 2 {
		t.Errorf("CLI removeProfile middle: expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("CLI removeProfile middle: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	foundFirst := false
	foundLast := false
	for _, p := range cfg.Profiles {
		if p.Name == "first-profile" {
			foundFirst = true
		}
		if p.Name == "last-profile" {
			foundLast = true
		}
	}
	if !foundFirst {
		t.Errorf("CLI removeProfile middle: expected to find 'first-profile'")
	}
	if !foundLast {
		t.Errorf("CLI removeProfile middle: expected to find 'last-profile'")
	}
	
	// Test removing first profile
	if err := cli.Run([]string{"remove", "--name", "first-profile"}); err != nil {
		t.Errorf("CLI removeProfile first: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err = Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after remove: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("CLI removeProfile first: expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("CLI removeProfile first: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	if cfg.Profiles[0].Name != "last-profile" {
		t.Errorf("CLI removeProfile first: expected remaining profile 'last-profile', got %s", cfg.Profiles[0].Name)
	}
	
	// Test removing last profile
	if err := cli.Run([]string{"remove", "--name", "last-profile"}); err != nil {
		t.Errorf("CLI removeProfile last: unexpected error: %v", err)
	}
	
	// Verify config was updated
	cfg, err = Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after remove: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("CLI removeProfile last: expected 0 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("CLI removeProfile last: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	
	// Test removing with missing name flag
	if err := cli.Run([]string{"remove"}); err == nil {
		t.Errorf("CLI removeProfile missing name: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile name is required") {
			t.Errorf("CLI removeProfile missing name: unexpected error: %v", err)
		}
	}
	
	// Test removing non-existent profile
	if err := cli.Run([]string{"remove", "--name", "nonexistent"}); err == nil {
		t.Errorf("CLI removeProfile nonexistent: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("CLI removeProfile nonexistent: unexpected error: %v", err)
		}
	}
}

func TestCLI_run_invalid_command(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	
	// Create empty config
	if err := CreateDefaultConfig(configPath); err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}
	
	cli := NewCLI(configPath)
	
	// Test invalid command
	if err := cli.Run([]string{"invalid-command"}); err == nil {
		t.Errorf("CLI invalid command: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("CLI invalid command: unexpected error: %v", err)
		}
	}
	
	// Test insufficient arguments (no subcommand)
	if err := cli.Run([]string{}); err == nil {
		t.Errorf("CLI insufficient arguments: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "usage: s3peep profile <command>") {
			t.Errorf("CLI insufficient arguments: unexpected error: %v", err)
		}
	}
}

// Helper function to capture stdout/stderr from a function
func captureOutput(t *testing.T, f func() error) string {
	// Redirect stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	
	// Run the function
	err := f()
	
	// Restore stdout and stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	
	// Read the output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	r.Close()
	
	if err != nil {
		t.Fatalf("Function returned error: %v", err)
	}
	
	return string(buf[:n])
}