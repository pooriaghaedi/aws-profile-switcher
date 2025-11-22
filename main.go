package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

const version = "1.0.0"

var (
	showVersion = flag.Bool("version", false, "Show version information")
	dryRun      = flag.Bool("dry-run", false, "Preview changes without applying them")
	showHelp    = flag.Bool("help", false, "Show help information")
)

func getFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Joining the home directory path with the path of the file
	filePath := filepath.Join(home, ".aws/credentials")
	return filePath
}

func loadFile() (*ini.File, error) {
	// Load the INI file
	cfg, err := ini.Load(getFilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS credentials file: %w", err)
	}
	return cfg, nil
}

func getProfiles(ini *ini.File) ([]string, error) {
	var profiles []string

	for _, section := range ini.Sections() {
		if section.Name() == "default" && section.Key("created_by_go").String() == "" {
			return nil, fmt.Errorf("found a [default] profile not created by this tool.\n\nTo fix this:\n  1. Rename your [default] profile to something like [default-backup] in ~/.aws/credentials\n  2. Run this tool again to create a new managed [default] profile")
		} else if section.Name() != "DEFAULT" && (section.Name() != "default" && section.Key("created_by_go").String() != "true") {
			profiles = append(profiles, section.Name())
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no AWS profiles found in ~/.aws/credentials")
	}

	return profiles, nil
}

func chooseProfile(profiles []string) (string, error) {
	// Create a new promptui Select
	prompt := promptui.Select{
		Label: "Select Profile",
		Items: profiles,
	}

	// Show the prompt and get the result
	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("profile selection failed: %w", err)
	}
	return result, nil
}

func updateDefaultProfile(ini *ini.File, profile string, isDryRun bool) error {
	return updateDefaultProfileWithPath(ini, profile, isDryRun, getFilePath())
}

func updateDefaultProfileWithPath(ini *ini.File, profile string, isDryRun bool, filePath string) error {
	section := ini.Section(profile)

	// Get the access key ID and secret access key for the selected profile
	accessKeyID := section.Key("aws_access_key_id").String()
	secretAccessKey := section.Key("aws_secret_access_key").String()
	region := section.Key("region").String()

	// Validate that the profile has required credentials
	if accessKeyID == "" || secretAccessKey == "" {
		return fmt.Errorf("profile '%s' is missing required credentials (aws_access_key_id or aws_secret_access_key)", profile)
	}

	if isDryRun {
		fmt.Println("\n--- DRY RUN MODE (no changes will be made) ---")
		fmt.Printf("Would update [default] profile with credentials from [%s]:\n", profile)
		fmt.Printf("  aws_access_key_id = %s\n", maskCredential(accessKeyID))
		fmt.Printf("  aws_secret_access_key = %s\n", maskCredential(secretAccessKey))
		if region != "" {
			fmt.Printf("  region = %s\n", region)
		}
		return nil
	}

	// Set these credentials for the default profile
	defaultSection := ini.Section("default")
	defaultSection.Key("aws_access_key_id").SetValue(accessKeyID)
	defaultSection.Key("aws_secret_access_key").SetValue(secretAccessKey)
	defaultSection.Key("created_by_go").SetValue("true")
	if region != "" {
		defaultSection.Key("region").SetValue(region)
	}

	// Save the INI file
	if err := ini.SaveTo(filePath); err != nil {
		return fmt.Errorf("failed to save credentials file: %w", err)
	}

	return nil
}

func maskCredential(cred string) string {
	if len(cred) <= 8 {
		return "****"
	}
	return cred[:4] + "****" + cred[len(cred)-4:]
}

func printHelp() {
	fmt.Printf(`AWS Profile Switcher v%s

USAGE:
  aws-switcher [OPTIONS]

DESCRIPTION:
  Interactive CLI tool to switch between AWS credential profiles.
  Updates the [default] profile in ~/.aws/credentials with credentials
  from your selected profile.

OPTIONS:
  --help      Show this help message
  --version   Show version information
  --dry-run   Preview changes without applying them

EXAMPLES:
  aws-switcher              Run interactively
  aws-switcher --dry-run    Preview what would change
  aws-switcher --version    Show version

For more information, visit: https://github.com/poori/Go-switcher
`, version)
}

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("AWS Profile Switcher v%s\n", version)
		os.Exit(0)
	}

	cfg, err := loadFile()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	profiles, err := getProfiles(cfg)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	result, err := chooseProfile(profiles)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if err := updateDefaultProfile(cfg, result, *dryRun); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if !*dryRun {
		fmt.Printf("\n✓ Successfully switched to profile: %s\n", result)
	}
}
