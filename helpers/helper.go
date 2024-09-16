package helpers

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// zeroBytes overwrites a byte slice with zeros.
func ZeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// equalBytes compares two byte slices for equality.
func EqualBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// PromptPassphrase prompts the user to enter and confirm a passphrase.
func PromptPassphrase(confirm bool) ([]byte, error) {
	fmt.Print("Enter passphrase: ")
	passphrase, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	fmt.Println()

	if confirm {
		fmt.Print("Confirm passphrase: ")
		confirmPassphrase, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, err
		}
		fmt.Println()

		if !EqualBytes(passphrase, confirmPassphrase) {
			return nil, errors.New("passphrases do not match")
		}
		// Zero out the confirmPassphrase
		ZeroBytes(confirmPassphrase)
	}

	return passphrase, nil
}
