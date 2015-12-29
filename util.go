package main

import (
	"errors"
	"fmt"
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

func bashPath() (string, error) {
	homeDir, err := homeDirectory()
	if err != nil {
		return "", err
	}

	// Check if there is a .bashrc
	bashrc := filepath.Join(homeDir, ".bashrc")
	if fileExists(bashrc) {
		return bashrc, nil
	}

	// Check if there is a .bash_profile
	bashProfile := filepath.Join(homeDir, ".bash_profile")
	if fileExists(bashProfile) {
		return bashProfile, nil
	}

	return "", errors.New("Couldn't find .bashrc or .bash_profile")
}

func checkFileForString(path string, s string) bool {
	// read .bashrc to make sure it hasn't been installed
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}

	lines := strings.Split(string(input), "\n")

	for _, line := range lines {
		if strings.Contains(line, s) {
			return true
		}
	}

	return false
}

func installed() bool {
	// Check if there is a .bashrc
	path, err := bashPath()
	if err != nil {
		return false // Doesn't have bash_profile or bashrc
	}

	if checkFileForString(path, rSourceName) {
		return true
	}

	return false
}

func sourceR(path string) error {
	// Get home directory
	homeDir, err := homeDirectory()
	if err != nil {
		return err
	}

	// Create .r.sh file in homeDirectory
	f, err := os.Create(filepath.Join(homeDir, rSourceName))
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(rBashFile)

	// Source .r.sh in bashrc
	bashFile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer bashFile.Close()

	rSourceFile := fmt.Sprintf("\n# r sourced from r -install \n. %s/%s", homeDir, rSourceName)
	if _, err = bashFile.WriteString(rSourceFile); err != nil {
		return err
	}

	// fmt.Printf("Installed %s to: %s \n", rSourceName, bashrc)
	fmt.Println("r successfully installed! Restart your bash shell.")
	return nil
}
