package config

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func CreateDirectoryIfNotExists(parentDir string) error {
	kuskConfigDirPath := path.Join(parentDir, ".kusk")
	f, err := os.Stat(kuskConfigDirPath)

	// if exists and is a directory, return
	if err == nil {
		if f.IsDir() {
			return nil
		}
		return fmt.Errorf("%s/.kusk exists but is a file. A directory was expected", parentDir)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to check for existence of %s/.kusk: %w", parentDir, err)
	}

	if err := os.Mkdir(kuskConfigDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create config directory at path %s: %w", kuskConfigDirPath, err)
	}

	return nil
}
