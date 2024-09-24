package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/fatih/color"
)

const (
	oldUsername        = "arch"
	loginUsername      = "login"
	skipFilename       = "/opt/startup/skip_bootscript"
	scriptFilename     = "/opt/startup/script.sh"
	testAddress        = "https://google.com"
	versionEnvKey      = "ARCHAVISTO_VERSION"
	privateTokenEnvKey = "GITLAB_TOKEN"
	repoURL            = "https://versioning.advans-group.com/api/v4/projects/1495/repository/files/packagesProfiles.json?ref=main"
	scriptTemplate     = `
#!/usr/bin/sh
set -o errexit
paru -Syu --skipreview
{{ if .Packages }}
paru -S {{ range .Packages }}{{ . }} {{ end }}
{{ end }}
{{ if ne .NewUsername .OldUsername }}
sudo usermod --login={{ .NewUsername }} --move-home --home=/home/{{ .NewUsername }} {{ .OldUsername }}
{{ end }}
sudo sed -i "s/{{ .LoginUsername }}/{{ .NewUsername }}/g" /etc/wsl.conf
touch {{ .SkipFile }}
`
)

type templateData struct {
	LoginUsername string
	OldUsername   string
	NewUsername   string
	SkipFile      string
	Packages      []string
}

type jsonData struct {
	Common   []string            `json:"common"`
	Profiles map[string][]string `json:"profiles"`
}

func main() {
	// checks if the skip file exists
	if _, err := os.Stat(skipFilename); err == nil {
		color.Yellow("Skipping script execution due to skip file %s,\n", skipFilename)
		os.Exit(1)
	}

	color.Cyan("Welcome to ArchAvisto!\n")

	version := os.Getenv(versionEnvKey)
	if version != "" {
		color.Yellow("Version: %s\n", version)
	}

	fmt.Println("Checking network connectivity...")
	if _, err := http.Get(testAddress); err != nil {
		color.Red("Unable to join Internet. Check the Confluence page for troobleshooting or Contact a DevOps internal member by Teams or by email devops-support@advans-group.atlassian.net")
		os.Exit(42)
	}

	fmt.Println("Network OK \u2714")

	updatePrompt()

	// Ask for the new username
	newUsername := userNamePrompt(oldUsername)

	// Fetching and parsing json file
	privateToken := os.Getenv(privateTokenEnvKey)
	if privateToken == "" {
		color.Red("Missing the GITLAB_TOKEN environment variable, cannot fetch the packages json file. Exiting...")
		os.Exit(1)
	}
	content, err := fetchJsonFiles(repoURL, privateToken)
	if err != nil {
		color.Red("Error while fetching the packages json file, it may be an internal Gitlab issue, please contact a DevOps internal member or IT for support: %s", err)
		os.Exit(1)
	}
	jsonData, err := parseJsonFile(content)
	if err != nil {
		color.Red("Error while parsing the packages json file: %s", err)
		os.Exit(1)
	}

	// Get the profiles
	var profiles []string
	for profile := range jsonData.Profiles {
		profiles = append(profiles, profile)
	}

	commonPackages := jsonData.Common

	// Ask for the profile and set packages list accordingly
	chosenProfile := profilePrompt(profiles)
	packageList := append(commonPackages, jsonData.Profiles[chosenProfile]...)

	// Prompt package list
	packageToInstall := packagesPrompt(packageList)

	// End of interactive prompts
	data := templateData{
		LoginUsername: loginUsername,
		OldUsername:   oldUsername,
		NewUsername:   newUsername,
		SkipFile:      skipFilename,
		Packages:      packageToInstall,
	}

	tmpl, err := template.New("example").Parse(scriptTemplate)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(scriptFilename, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}
}
