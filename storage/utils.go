package storage

import (
	"os"
)

// createDirIfNotExists creates a directory if it doesn't exist
func createDirIfNotExists(dir string) error {
	// Check if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create the directory with permissions
		return os.MkdirAll(dir, 0755)
	} else if err != nil {
		return err
	}
	return nil
}
