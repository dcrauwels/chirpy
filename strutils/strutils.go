package strutils

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"
)

func ChirpLength(chirp string, maxLength int) error {
	if len(chirp) > maxLength {
		return fmt.Errorf("chirp exceeds maximum length of %d characters", maxLength)
	}
	return nil
}

func ReplaceWord(input string, target []string, fixed string) string {
	slicedInput := strings.Split(input, " ")
	fixedInput := make([]string, len(slicedInput))
	for i, w := range slicedInput {
		if slices.Contains(target, strings.ToLower(w)) {
			fixedInput[i] = fixed
		} else {
			fixedInput[i] = w
		}
	}
	output := strings.Join(fixedInput, " ")
	return output
}

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}
