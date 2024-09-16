package main

import (
	"errors"
	"os"
	"regexp"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/multichoose"
	"github.com/fatih/color"
)

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
