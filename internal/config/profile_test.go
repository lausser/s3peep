package config

import (
	"strings"
	"testing"
)

func TestProfileValidate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: Profile{
				Name:            "test",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			profile: Profile{
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
			wantErr: true,
		},
		{
			name: "missing region",
			profile: Profile{
				Name:            "test",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
			wantErr: true,
		},
		{
			name: "missing access key",
			profile: Profile{
				Name:            "test",
				Region:          "us-east-1",
				SecretAccessKey: "testsecret",
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			profile: Profile{
				Name:            "test",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("ProfileValidate() error = %v, wantErr %t", err, tt.wantErr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ProfileValidate() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				// Just checking that we got an error, not the specific message
			}
		})
	}
}

func TestAddProfile(t *testing.T) {
	cfg := &Config{}
	
	// Test adding first profile
	profile := Profile{
		Name:            "test1",
		Region:          "us-east-1",
		AccessKeyID:     "testkey1",
		SecretAccessKey: "testsecret1",
	}
	
	if err := AddProfile(cfg, profile); err != nil {
		t.Errorf("AddProfile first profile: unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("AddProfile first profile: expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "test1" {
		t.Errorf("AddProfile first profile: expected ActiveProfile 'test1', got %s", cfg.ActiveProfile)
	}
	
	// Test adding second profile
	profile2 := Profile{
		Name:            "test2",
		Region:          "us-west-2",
		AccessKeyID:     "testkey2",
		SecretAccessKey: "testsecret2",
	}
	
	if err := AddProfile(cfg, profile2); err != nil {
		t.Errorf("AddProfile second profile: unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 2 {
		t.Errorf("AddProfile second profile: expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "test2" {
		t.Errorf("AddProfile second profile: expected ActiveProfile 'test2', got %s", cfg.ActiveProfile)
	}
	
	// Test adding duplicate profile
	if err := AddProfile(cfg, profile); err == nil {
		t.Errorf("AddProfile duplicate profile: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile already exists") {
			t.Errorf("AddProfile duplicate profile: unexpected error: %v", err)
		}
	}
}

func TestRemoveProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{
				Name: "profile1",
			},
			{
				Name: "profile2",
			},
			{
				Name: "profile3",
			},
		},
		ActiveProfile: "profile2",
	}
	
	// Test removing middle profile
	if err := RemoveProfile(cfg, "profile2"); err != nil {
		t.Errorf("RemoveProfile middle profile: unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 2 {
		t.Errorf("RemoveProfile middle profile: expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("RemoveProfile middle profile: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	
	// Test removing first profile
	if err := RemoveProfile(cfg, "profile1"); err != nil {
		t.Errorf("RemoveProfile first profile: unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("RemoveProfile first profile: expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("RemoveProfile first profile: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	
	// Test removing last profile
	if err := RemoveProfile(cfg, "profile3"); err != nil {
		t.Errorf("RemoveProfile last profile: unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("RemoveProfile last profile: expected 0 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("RemoveProfile last profile: expected ActiveProfile '', got %s", cfg.ActiveProfile)
	}
	
	// Test removing non-existent profile
	if err := RemoveProfile(cfg, "nonexistent"); err == nil {
		t.Errorf("RemoveProfile non-existent profile: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("RemoveProfile non-existent profile: unexpected error: %v", err)
		}
	}
}

func TestSwitchProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{
				Name: "profile1",
			},
			{
				Name: "profile2",
			},
			{
				Name: "profile3",
			},
		},
		ActiveProfile: "profile1",
	}
	
	// Test switching to existing profile
	if err := SwitchProfile(cfg, "profile3"); err != nil {
		t.Errorf("SwitchProfile to existing profile: unexpected error: %v", err)
	}
	if cfg.ActiveProfile != "profile3" {
		t.Errorf("SwitchProfile to existing profile: expected ActiveProfile 'profile3', got %s", cfg.ActiveProfile)
	}
	
	// Test switching to another existing profile
	if err := SwitchProfile(cfg, "profile2"); err != nil {
		t.Errorf("SwitchProfile to another existing profile: unexpected error: %v", err)
	}
	if cfg.ActiveProfile != "profile2" {
		t.Errorf("SwitchProfile to another existing profile: expected ActiveProfile 'profile2', got %s", cfg.ActiveProfile)
	}
	
	// Test switching to non-existent profile
	if err := SwitchProfile(cfg, "nonexistent"); err == nil {
		t.Errorf("SwitchProfile non-existent profile: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("SwitchProfile non-existent profile: unexpected error: %v", err)
		}
	}

	// Test switching with empty profiles list
	cfg.Profiles = []Profile{}
	if err := SwitchProfile(cfg, "any"); err == nil {
		t.Errorf("SwitchProfile with empty profiles: expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "profile not found") {
			t.Errorf("SwitchProfile with empty profiles: unexpected error: %v", err)
		}
	}
}

func TestUpdateProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{
				Name:            "profile1",
				Region:          "us-east-1",
				AccessKeyID:     "old-access-key",
				SecretAccessKey: "old-secret-key",
				EndpointURL:     "http://old-endpoint",
				Bucket:          "old-bucket",
			},
			{
				Name:            "profile2",
				Region:          "eu-west-1",
				AccessKeyID:     "key2",
				SecretAccessKey: "secret2",
				EndpointURL:     "",
				Bucket:          "",
			},
		},
	}

	// Test updating region
	updates := Profile{
		Region: "us-west-2",
	}
	if err := UpdateProfile(cfg, "profile1", updates); err != nil {
		t.Errorf("UpdateProfile region: unexpected error: %v", err)
	}
	if cfg.Profiles[0].Region != "us-west-2" {
		t.Errorf("UpdateProfile region: expected 'us-west-2', got %s", cfg.Profiles[0].Region)
	}
	// Ensure other fields are unchanged
	if cfg.Profiles[0].AccessKeyID != "old-access-key" {
		t.Errorf("UpdateProfile region: access key should not change")
	}

	// Test updating endpoint
	updates = Profile{
		EndpointURL: "http://new-endpoint",
	}
	if err := UpdateProfile(cfg, "profile1", updates); err != nil {
		t.Errorf("UpdateProfile endpoint: unexpected error: %v", err)
	}
	if cfg.Profiles[0].EndpointURL != "http://new-endpoint" {
		t.Errorf("UpdateProfile endpoint: expected 'http://new-endpoint', got %s", cfg.Profiles[0].EndpointURL)
	}

	// Test updating access key and secret
	updates = Profile{
		AccessKeyID:     "new-access-key",
		SecretAccessKey: "new-secret-key",
	}
	if err := UpdateProfile(cfg, "profile1", updates); err != nil {
		t.Errorf("UpdateProfile credentials: unexpected error: %v", err)
	}
	if cfg.Profiles[0].AccessKeyID != "new-access-key" {
		t.Errorf("UpdateProfile credentials: access key not updated")
	}
	if cfg.Profiles[0].SecretAccessKey != "new-secret-key" {
		t.Errorf("UpdateProfile credentials: secret key not updated")
	}

	// Test updating bucket
	updates = Profile{
		Bucket: "new-bucket",
	}
	if err := UpdateProfile(cfg, "profile1", updates); err != nil {
		t.Errorf("UpdateProfile bucket: unexpected error: %v", err)
	}
	if cfg.Profiles[0].Bucket != "new-bucket" {
		t.Errorf("UpdateProfile bucket: expected 'new-bucket', got %s", cfg.Profiles[0].Bucket)
	}

	// Test updating non-existent profile
	updates = Profile{
		Region: "ap-south-1",
	}
	if err := UpdateProfile(cfg, "nonexistent", updates); err == nil {
		t.Errorf("UpdateProfile non-existent: expected error, got nil")
	} else if !strings.Contains(err.Error(), "profile not found") {
		t.Errorf("UpdateProfile non-existent: unexpected error: %v", err)
	}

	// Test updating second profile (to ensure we don't always update first)
	updates = Profile{
		Region: "eu-central-1",
	}
	if err := UpdateProfile(cfg, "profile2", updates); err != nil {
		t.Errorf("UpdateProfile second profile: unexpected error: %v", err)
	}
	if cfg.Profiles[1].Region != "eu-central-1" {
		t.Errorf("UpdateProfile second profile: region not updated")
	}
	// First profile should be unchanged
	if cfg.Profiles[0].Region != "us-west-2" {
		t.Errorf("UpdateProfile second profile: first profile was incorrectly modified")
	}
}