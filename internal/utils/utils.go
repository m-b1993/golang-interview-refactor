package utils

import (
	"os"
	"path/filepath"
)

func GetRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, "..", "..")
}

func GetConfigDir() string {
	rootDir := GetRootDir()
	return filepath.Join(rootDir, "config")
}

func GetTemplatesDir() string {
	rootDir := GetRootDir()
	return filepath.Join(rootDir, "static", "templates")
}
