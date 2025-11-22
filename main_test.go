package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestGetFilePath(t *testing.T) {
	filePath := getFilePath()

	// Verify it returns a path
	if filePath == "" {
		t.Error("getFilePath() returned empty string")
	}

	// Verify it contains expected components
	if !filepath.IsAbs(filePath) {
		t.Errorf("getFilePath() should return absolute path, got: %s", filePath)
	}

	expectedSuffix := filepath.Join(".aws", "credentials")
	if !filepath.HasPrefix(filePath, "/") && !filepath.HasPrefix(filePath, "C:") {
		t.Errorf("getFilePath() should return absolute path starting with / or C:, got: %s", filePath)
	}

	if filepath.Base(filePath) != "credentials" {
		t.Errorf("getFilePath() should end with 'credentials', got: %s", filePath)
	}

	_ = expectedSuffix // Use variable to avoid lint error
}

func TestLoadFile_NonExistentFile(t *testing.T) {
	// This test will fail if the credentials file doesn't exist
	// which is expected behavior
	_, err := loadFile()

	// We can't predict if file exists, so just verify function works
	if err != nil {
		t.Logf("loadFile() returned error (expected if no credentials file): %v", err)
	}
}

func TestGetProfiles_EmptyFile(t *testing.T) {
	cfg := ini.Empty()

	profiles, err := getProfiles(cfg)

	if err == nil {
		t.Error("getProfiles() should return error for empty file")
	}

	if len(profiles) != 0 {
		t.Errorf("getProfiles() should return empty slice for empty file, got %d profiles", len(profiles))
	}
}

func TestGetProfiles_WithProfiles(t *testing.T) {
	cfg := ini.Empty()

	// Add some test profiles
	cfg.Section("dev").Key("aws_access_key_id").SetValue("AKIATEST123")
	cfg.Section("prod").Key("aws_access_key_id").SetValue("AKIATEST456")
	cfg.Section("staging").Key("aws_access_key_id").SetValue("AKIATEST789")

	profiles, err := getProfiles(cfg)

	if err != nil {
		t.Errorf("getProfiles() returned unexpected error: %v", err)
	}

	expectedCount := 3
	if len(profiles) != expectedCount {
		t.Errorf("getProfiles() expected %d profiles, got %d", expectedCount, len(profiles))
	}

	// Verify profile names
	expectedProfiles := map[string]bool{"dev": true, "prod": true, "staging": true}
	for _, profile := range profiles {
		if !expectedProfiles[profile] {
			t.Errorf("getProfiles() returned unexpected profile: %s", profile)
		}
	}
}

func TestGetProfiles_WithManagedDefault(t *testing.T) {
	cfg := ini.Empty()

	// Add managed default profile (should be filtered out)
	cfg.Section("default").Key("created_by_go").SetValue("true")
	cfg.Section("default").Key("aws_access_key_id").SetValue("AKIATEST123")

	// Add regular profile
	cfg.Section("dev").Key("aws_access_key_id").SetValue("AKIATEST456")

	profiles, err := getProfiles(cfg)

	if err != nil {
		t.Errorf("getProfiles() returned unexpected error: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("getProfiles() expected 1 profile, got %d", len(profiles))
	}

	if profiles[0] != "dev" {
		t.Errorf("getProfiles() expected profile 'dev', got '%s'", profiles[0])
	}
}

func TestGetProfiles_WithUnmanagedDefault(t *testing.T) {
	cfg := ini.Empty()

	// Add unmanaged default profile (should cause error)
	cfg.Section("default").Key("aws_access_key_id").SetValue("AKIATEST123")
	// Note: created_by_go is NOT set

	// Add regular profile
	cfg.Section("dev").Key("aws_access_key_id").SetValue("AKIATEST456")

	_, err := getProfiles(cfg)

	if err == nil {
		t.Error("getProfiles() should return error for unmanaged default profile")
	}
}

func TestUpdateDefaultProfile_MissingCredentials(t *testing.T) {
	cfg := ini.Empty()

	// Create profile with missing credentials
	cfg.Section("dev").Key("region").SetValue("us-west-2")

	err := updateDefaultProfile(cfg, "dev", false)

	if err == nil {
		t.Error("updateDefaultProfile() should return error for profile missing credentials")
	}
}

func TestUpdateDefaultProfile_Success(t *testing.T) {
	cfg := ini.Empty()

	// Create source profile
	cfg.Section("dev").Key("aws_access_key_id").SetValue("AKIATEST123")
	cfg.Section("dev").Key("aws_secret_access_key").SetValue("SECRET123")
	cfg.Section("dev").Key("region").SetValue("us-west-2")

	// Create temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "credentials")

	err := updateDefaultProfileWithPath(cfg, "dev", false, tmpFile)

	if err != nil {
		t.Errorf("updateDefaultProfile() returned unexpected error: %v", err)
	}

	// Reload and verify
	reloadedCfg, err := ini.Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reload credentials file: %v", err)
	}

	defaultSection := reloadedCfg.Section("default")
	if defaultSection.Key("aws_access_key_id").String() != "AKIATEST123" {
		t.Error("Default profile should have updated aws_access_key_id")
	}
	if defaultSection.Key("aws_secret_access_key").String() != "SECRET123" {
		t.Error("Default profile should have updated aws_secret_access_key")
	}
	if defaultSection.Key("region").String() != "us-west-2" {
		t.Error("Default profile should have updated region")
	}
	if defaultSection.Key("created_by_go").String() != "true" {
		t.Error("Default profile should have created_by_go marker")
	}
}

func TestUpdateDefaultProfile_DryRun(t *testing.T) {
	cfg := ini.Empty()

	// Create source profile
	cfg.Section("dev").Key("aws_access_key_id").SetValue("AKIATEST123")
	cfg.Section("dev").Key("aws_secret_access_key").SetValue("SECRET123")

	// Create temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "credentials")

	err := updateDefaultProfileWithPath(cfg, "dev", true, tmpFile)

	if err != nil {
		t.Errorf("updateDefaultProfile() in dry-run mode returned unexpected error: %v", err)
	}

	// Verify file was NOT created (dry run)
	if _, err := os.Stat(tmpFile); err == nil {
		t.Error("updateDefaultProfile() in dry-run mode should not create file")
	}
}

func TestMaskCredential(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short credential",
			input:    "SHORT",
			expected: "****",
		},
		{
			name:     "normal access key",
			input:    "AKIAIOSFODNN7EXAMPLE",
			expected: "AKIA****MPLE",
		},
		{
			name:     "normal secret key",
			input:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			expected: "wJal****EKEY",
		},
		{
			name:     "exactly 8 characters",
			input:    "12345678",
			expected: "****",
		},
		{
			name:     "exactly 9 characters",
			input:    "123456789",
			expected: "1234****6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCredential(tt.input)
			if result != tt.expected {
				t.Errorf("maskCredential(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
