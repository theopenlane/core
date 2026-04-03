package fgaversion

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const packageName = "github.com/openfga/openfga"

func GetVersion() (string, error) {
	goModPath, err := findGoModPath()
	if err != nil {
		return "", err
	}

	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, packageName) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
		}
	}
	return "", scanner.Err()
}

func findGoModPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return modPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return "", os.ErrNotExist
}
