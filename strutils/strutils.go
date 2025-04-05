package strutils

import "fmt"

func ChirpLength(chirp string, maxLength int) error {
	if len(chirp) > maxLength {
		return fmt.Errorf("chirp exceeds maximum length of %d characters", maxLength)
	}
	return nil
}
