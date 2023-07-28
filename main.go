package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
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

func loadFile() *ini.File {
	// Load the INI file
	cfg, err := ini.Load(getFilePath())
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func getProfiles(ini *ini.File) []string {
	var profiles []string

	for _, section := range ini.Sections() {
		if section.Name() == "default" && section.Key("created_by_go").String() == "" {
			panic("Please change default profile to something else")
		} else if section.Name() != "DEFAULT" && (section.Name() != "default" && section.Key("created_by_go").String() != "true") {
			profiles = append(profiles, section.Name())
		}
	}
	return profiles
}

func chooseProfile(profiles []string) string {
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
	return result
}

func updateDefaultProfile(ini *ini.File, profile string) {
	// fmt.Println(profile)
	section := ini.Section(profile)

	// Get the access key ID and secret access key for the selected profile
	accessKeyID := section.Key("aws_access_key_id").String()
	secretAccessKey := section.Key("aws_secret_access_key").String()
	region := section.Key("region").String()

	// Set these credentials for the default profile
	defaultSection := ini.Section("default")
	defaultSection.Key("aws_access_key_id").SetValue(accessKeyID)
	defaultSection.Key("aws_secret_access_key").SetValue(secretAccessKey)
	defaultSection.Key("created_by_go").SetValue("true")
	if region != "" {
		defaultSection.Key("region").SetValue(region)
	}

	// Save the INI file
	ini.SaveTo(getFilePath())
}

func main() {
	cfg := loadFile()
	profiles := getProfiles(cfg)
	result := chooseProfile(profiles)

	updateDefaultProfile(cfg, result)
}
