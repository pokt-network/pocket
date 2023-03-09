package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Check directory exists and creates path if it doesn't exist
func DirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		// Exists but not directory
		if !stat.IsDir() {
			return false, fmt.Errorf("Path exists but is not a directory: %s", path)
		}
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		// Create directories in path recursively
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return false, fmt.Errorf("Error creating directory at path: %s, (%v)", path, err.Error())
		}
		return true, nil
	}
	return false, err
}

// Check file at the given path exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func WriteOutput(msg any, outputFile string) error {
	if outputFile == "" {
		fmt.Println(msg)
		return nil
	}
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	switch m := msg.(type) {
	case string:
		if _, err := file.WriteString(m); err != nil {
			return err
		}
	case []byte:
		if _, err := file.Write(m); err != nil {
			return err
		}
	}
	return file.Close()
}

func ReadInput(inputFile string) ([]byte, error) {
	exists, err := FileExists(inputFile)
	if err != nil {
		return []byte{}, fmt.Errorf("Error checking input file: %v\n", err)
	}
	if !exists {
		return []byte{}, fmt.Errorf("Input file not found: %v\n", inputFile)
	}
	rawBz, err := os.ReadFile(inputFile)
	if err != nil {
		return []byte{}, err
	}
	return rawBz, nil
}
