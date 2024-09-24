package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/multichoose"
	"github.com/fatih/color"
)

type FileResponse struct {
	Content string `json:"content"` // Base64 encoded content
}

func checkError(err error, message string) {
	if err != nil {
		color.Red("%s: %s", message, err)
		os.Exit(1)
	}
}

func updatePrompt() {
	result, err := prompt.New().Ask(" Do you agree to update the system packages").Choose([]string{"Yes", "No"})
	checkError(err, "Prompt failed")

	if result == "No" {
		color.Red("Cannot continue without updating the system.")
		os.Exit(1)
	}
}

// Ensure that the username is Unix compliant
func validateUsername(input string) error {
	var validUsernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}[^-]$`)

	if !validUsernameRegex.MatchString(input) {
		return errors.New("invalid username")
	}
	return nil
}

func userNamePrompt(defaultUsername string) string {
	name, err := prompt.New().Ask(" What should I call you ?").Input(defaultUsername)
	checkError(err, "Prompt failed")
	if validateUsername(name) != nil {
		color.Red("Invalid username! your username must be Unix compliant, please try again.")
		name = userNamePrompt(defaultUsername)
	} else if name == "" {
		return defaultUsername
	}
	return name
}

func fetchJsonFiles(url string, privateToken string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Private-Token", privateToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != 200 {
		color.Red("Failed to fetch packages json file: %s", resp.Status)
		return "", errors.New("Failed to fetch packages json file")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var fileresp FileResponse
	error := json.Unmarshal(body, &fileresp)
	if error != nil {
		return "", err
	}
	// Decoding file content
	fileContent, err := base64.StdEncoding.DecodeString(fileresp.Content)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

func parseJsonFile(jsonFile string) (jsonData, error) {
	var data jsonData
	err := json.Unmarshal([]byte(jsonFile), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func profilePrompt(profiles []string) string {
	result, err := prompt.New().Ask(" Choose a profile").Choose(profiles)
	checkError(err, "Prompt failed")
	return result
}

func packagesPrompt(packages []string) []string {
	result, err := prompt.New().Ask(" Select the packages you want to install").MultiChoose(
		packages,
		multichoose.WithTheme(multichoose.ThemeDot),
	)
	checkError(err, "Prompt failed")
	return result
}
