package renv

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var envMode = flag.String("envMode", "", "env mode")
var ParseAtLocationParam = flag.String("parse_at_location", "", "parse at location")

func ParseCmd(v interface{}) {
	Parse(*envMode, *ParseAtLocationParam, v)
}

func Parse(env, parseAtLocation string, v interface{}) {
	var fileName string
	if env == "" {
		fileName = ".env.local.yaml"
	} else {
		fileName = fmt.Sprintf(".env.%s.yaml", env)
	}

	// If specific location provided, use it
	if parseAtLocation != "" {
		foundPath := filepath.Join(parseAtLocation, fileName)
		if _, err := os.Stat(foundPath); os.IsNotExist(err) {
			panic(fmt.Sprintf("missing env file: %s", foundPath))
		}
		ParseAtLocation(foundPath, v)
		return
	}

	// Try to find the env file in multiple locations
	searchPaths := []string{
		fileName,                          // Current directory
		filepath.Join("server", fileName), // From project root
	}

	// Also try relative to executable location
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		// If executable is in server/, look for env file there
		searchPaths = append(searchPaths, filepath.Join(execDir, fileName))
		// If executable is in $GOPATH/bin, look in common project locations
		searchPaths = append(searchPaths, filepath.Join(execDir, "..", "src", "elotus_test", "server", fileName))
	}

	// Get current working directory for better error message
	cwd, _ := os.Getwd()

	var foundPath string
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		panic(fmt.Sprintf("missing env file %s\nCurrent directory: %s\nSearched paths: %s\nTip: Run from project root or use -parse_at_location flag",
			fileName, cwd, strings.Join(searchPaths, ", ")))
	}

	ParseAtLocation(foundPath, v)
}

func ParseAtLocation(fileName string, v interface{}) {
	raw, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(raw, v)
	if err != nil {
		panic(err)
	}
}
