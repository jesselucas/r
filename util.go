package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func homeDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

func bashrcPath() (string, error) {
	homeDir, err := homeDirectory()
	if err != nil {
		return "", err
	}

	// Check if there is a .bashrc
	bashrc := filepath.Join(homeDir, ".bashrc")
	if fileExists(bashrc) {
		return bashrc, nil
	}

	return "", errors.New("Couldn't find .bashrc")
}

func installed() bool {
	// Check if there is a .bashrc
	bashrc, err := bashrcPath()
	if err != nil {
		return false
	}

	// read .bashrc to make sure it hasn't been installed
	input, err := ioutil.ReadFile(bashrc)
	if err != nil {
		return false
	}

	lines := strings.Split(string(input), "\n")

	for _, line := range lines {
		if strings.Contains(line, rSourceName) {
			return true
		}
	}

	return false
}
