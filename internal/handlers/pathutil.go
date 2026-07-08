package handlers

import (
	"errors"
	"path/filepath"
	"strings"
)

// cleanHandlerFilename strips directory components and rejects empty/parent names.
func cleanHandlerFilename(name string) (string, error) {
	cleanName := filepath.Clean(filepath.Base(name))
	if cleanName == "." || cleanName == ".." || cleanName == string(filepath.Separator) {
		return "", errors.New("invalid filename")
	}
	return cleanName, nil
}

// resolveSafePath joins name under base and rejects any path that escapes base.
func resolveSafePath(base, name string) (string, error) {
	cleanName, err := cleanHandlerFilename(name)
	if err != nil {
		return "", err
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	return resolvePathWithinBase(absBase, cleanName)
}

func resolvePathWithinBase(absBase, cleanName string) (string, error) {
	absResolved, err := filepath.Abs(filepath.Join(absBase, cleanName))
	if err != nil {
		return "", err
	}
	basePrefix := absBase + string(filepath.Separator)
	if absResolved != absBase && !strings.HasPrefix(absResolved, basePrefix) {
		return "", errors.New("path escapes base directory")
	}
	return absResolved, nil
}
