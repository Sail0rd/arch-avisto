package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/fatih/color"
)

const (
	oldUsername    = "arch"
	loginUsername  = "login"
	skipFilename   = "/opt/startup/skip_bootscript"
	scriptFilename = "/opt/startup/script.sh"
	testAddress    = "https://google.com"
	versionEnvKey  = "ARCHAVISTO_VERSION"
	scriptTemplate = `
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
echo "Installation complete!"
echo "You can now go back to Powershell and start WSL using the command: wsl -u {{ .NewUsername }} -d <distro-name> OR wsl -t <distro-name> and then wsl -d <distro-name>"
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

var (
	commonPackages = []string{
		"duf", "dust", "helm", "helmfile", "jq", "k9s", "kubectl", "micro", "navi", "pre-commit", "skaffold", "unzip", "wget",
	}
	devPackages = []string{
		"gitleaks",
	}
	devOpsPackages = []string{
		"ansible", "bottom", "htop", "iperf", "gnu-netcat", "net-tools", "pgcli", "screen", "sshuttle", "tcpdump", "inetutils", "terraform", "tmux",
	}
	profiles = []string{"Dev", "DevOps"}
)

func main() {
	// checks if the skip file exists
	if _, err := os.Stat(skipFilename); err == nil {
		color.Yellow("Skipping script execution due to skip file \u1F680 \n")
		os.Exit(1)
	}

	color.Green("Welcome to ArchAvisto\n")

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

	// Ask for the profile and set packages list accordingly
	chosenProfile := profilePrompt(profiles)

	packageList := commonPackages
	if chosenProfile == "Dev" {
		packageList = append(packageList, devPackages...)
	} else {
		packageList = append(packageList, devOpsPackages...)
	}

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
