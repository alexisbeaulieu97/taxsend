package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

var promptSecret = defaultPromptSecret

func defaultPromptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	secret, err := term.ReadPassword(os.Stdin.Fd())
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	secretText := strings.TrimSpace(string(secret))
	if secretText == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return secretText, nil
}

func promptPassphrase(confirm bool) (string, error) {
	first, err := promptSecret("Passphrase: ")
	if err != nil {
		return "", err
	}
	if !confirm {
		return first, nil
	}
	second, err := promptSecret("Confirm passphrase: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("passphrases do not match")
	}
	return first, nil
}
