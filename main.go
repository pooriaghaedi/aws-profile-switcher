package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

func main() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Joining the home directory path with the path of the file
	filePath := filepath.Join(home, ".aws/credentials")

	// Load the INI file
	cfg, err := ini.Load(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var profiles []string
	for _, section := range cfg.Sections() {
		if section.Name() != "DEFAULT" {
			if section.Name() != "default" {
				profiles = append(profiles, section.Name())
			} else if section.Key("created_by_go").String() != "true" {
				fmt.Println("Change default profile name to something else")
				os.Exit(1)
			}
		}
	}

	// Create a new promptui Select
	prompt := promptui.Select{
		Label: "Select Profile",
		Items: profiles,
	}

	// Show the prompt and get the result
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	// Get the selected profile
	section := cfg.Section(result)

	// Get the access key ID and secret access key for the selected profile
	accessKeyID := section.Key("aws_access_key_id").String()
	secretAccessKey := section.Key("aws_secret_access_key").String()
	region := section.Key("region").String()

	// Set these credentials for the default profile
	defaultSection := cfg.Section("default")
	defaultSection.Key("aws_access_key_id").SetValue(accessKeyID)
	defaultSection.Key("aws_secret_access_key").SetValue(secretAccessKey)
	defaultSection.Key("created_by_go").SetValue("true")
	if region != "" {
		defaultSection.Key("region").SetValue(region)
	}
	// Save the INI file
	cfg.SaveTo(filePath)

}
