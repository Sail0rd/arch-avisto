package main

import (
	"net/http"
	"os"
	"text/template"

	"github.com/fatih/color"
)

const (
	oldUsername        = "arch"
	loginUsername      = "login"
	skipFilename       = "/opt/startup/skip_bootscript"
	bootscriptVersion  = "1.1.2"
	scriptFilename     = "/opt/startup/script.sh"
	testAddress        = "https://google.com"
	versionEnvKey      = "ARCHAVISTO_VERSION"
	maintainers        = "Yann Lacroix <yann.lacroix@avisto.com>, Mathis Guilbaud <mathis.guilbaud@avisto.com>"
	privateTokenEnvKey = "GITLAB_TOKEN"
	fileUrl            = "https://versioning.advans-group.com/api/v4/projects/1495/repository/files/packages.json?ref=main"
	scriptTemplate     = `
#!/usr/bin/sh
set -o errexit
# Bootscript version: {{ .BootscriptVersion }}
sudo -u {{ .OldUsername }} paru -Sy --skipreview archlinux-keyring
sudo -u {{ .OldUsername }} paru -Syu --skipreview
{{ if .Packages }}
sudo -u {{ .OldUsername }} paru -S {{ range .Packages }}{{ . }} {{ end }}
{{ end }}
{{ if ne .NewUsername .OldUsername }}
sudo usermod --login={{ .NewUsername }} --shell /usr/sbin/{{ .Shell }} --move-home --home=/home/{{ .NewUsername }} {{ .OldUsername }}
{{ else }}
sudo usermod --shell /usr/sbin/{{ .Shell }} {{ .OldUsername }}
{{ end }}
sudo sed -i "s/{{ .LoginUsername }}/{{ .NewUsername }}/g" /etc/wsl.conf
touch {{ .SkipFile }}
`
)

type templateData struct {
	LoginUsername     string
	OldUsername       string
	NewUsername       string
	SkipFile          string
	BootscriptVersion string
	Shell             string
	Packages          []string
}

// jsonData is the struct that represents the json file
// {"common": [{"duf": "fancy disk usage"}], "profiles": { "dev": [{"gitleaks": "detects secrets"}], "devops": [{"ansible": "automation"}] }}

type pkgData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonData struct {
	Common   []pkgData            `json:"common"`
	Profiles map[string][]pkgData `json:"profiles"`
}

func main() {
	ClearScreen()

	// checks if the skip file exists
	if _, err := os.Stat(skipFilename); err == nil {
		color.Yellow("Skipping script execution due to skip file %s,\n", skipFilename)
		// exit 1 is important otherwise the slogin script will continue its execution
		os.Exit(1)
	}

	color.Cyan("Welcome to ArchAvisto!\n")
	color.Green("Maintainers: %s", maintainers)

	if version := os.Getenv(versionEnvKey); version != "" {
		color.Yellow("Image Version: %s\n", version)
		color.Yellow("Bootscript Version: %s\n", bootscriptVersion)
	}

	color.Cyan("Checking network connectivity...")
	if _, err := http.Get(testAddress); err != nil {
		color.Red("Unable to join Internet.")
		os.Exit(42)
	}

	color.Green("Network OK \u2714")

	// Ask for the permission to update the system
	if updatePrompt() == "No" {
		color.Red("Cannot continue without updating the system packages.")
		os.Exit(1)
	}

	// Ask for the new username
	newUsername := userNamePrompt(oldUsername)

	// Fetching and parsing json file
	privateToken := os.Getenv(privateTokenEnvKey)
	if privateToken == "" {
		color.Red("Missing the GITLAB_TOKEN environment variable, cannot fetch the packages json file. Exiting...")
		os.Exit(1)
	}
	content, err := fetchJsonFiles(fileUrl, privateToken)
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

	var commonPackages []string
	var packageDescription []string
	for _, pkg := range jsonData.Common {
		commonPackages = append(commonPackages, pkg.Name)
		packageDescription = append(packageDescription, pkg.Description)
	}

	// Ask for the profile and set packages list accordingly
	chosenProfiles := profilePrompt(profiles)

	packageList := commonPackages
	for _, profile := range chosenProfiles {
		for _, pkg := range jsonData.Profiles[profile] {
			packageList = append(packageList, pkg.Name)
			packageDescription = append(packageDescription, pkg.Description)
		}
	}

	// Prompt package list
	packageToInstall := packagesPrompt(packageList, packageDescription)

	chosenShell := shellPrompt([]string{"bash", "fish", "zsh"})

	// End of interactive prompts
	data := templateData{
		LoginUsername:     loginUsername,
		OldUsername:       oldUsername,
		NewUsername:       newUsername,
		SkipFile:          skipFilename,
		BootscriptVersion: bootscriptVersion,
		Shell:             chosenShell,
		Packages:          packageToInstall,
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
