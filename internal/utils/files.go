package utils

import (
	"errors"
	"os"
)

func CheckIfFileExists(filename string) bool {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
