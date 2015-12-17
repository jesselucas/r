package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	// Setup flags
	initPtr := flag.Bool("init", false, "initializes `r` and starts tracking all commands used in bash")
	statsPtr := flag.Bool("stats", false, "show stats and usage of `r`")
	commandsPtr := flag.Bool("commands", false, "show all commands that `r` will track")
	addPtr := flag.String("add", "", "show stats and usage of `r`")
	flag.Parse()

	// Check if `init` flag is passed
	if *initPtr {
		initialize()
		os.Exit(0)
	}

	// Check if `stats` flag is passed
	if *statsPtr {
		stats()
		os.Exit(0)
	}

	// Check if `commands` flag is passed
	if *commandsPtr {
		commands, err := listCommands()
		if err != nil {
			log.Fatal(err)
		}

		for _, c := range commands {
			fmt.Println(c)
		}

		os.Exit(0)
	}

	// Check if `add` flag is passed
	if *addPtr != "" {
		args := strings.Split(*addPtr, ":")
		err := add(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}
	}
}

// initialize uses compgen -c to get a list of all bash commands
// then creates aliases for each of them to store usage and directory
func initialize() {
	fmt.Println("initialize")

	err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	// commandList, err := listCommands()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = createAlias(commandList)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// setupDB verifies and creates boltDB in ~ home folder
func setupDB() error {
	// TODO setup boltdb
	fmt.Println("setup boltdb")

	return nil
}

// add checks if command being passed is in the listCommands
// then stores the command and workding directory
func add(path string, promptCmd string) error {
	// only want the first command in the promptCmd string
	cmd := strings.Split(promptCmd, " ")[0]

	// first check to see if cmd is valid
	commands, err := listCommands()
	if err != nil {
		return err
	}

	containsCmd := func() bool {
		for _, c := range commands {
			if c == cmd {
				return true
			}
		}
		return false
	}

	// check if the command is valid
	if !containsCmd() {
		return nil
	}

	fmt.Printf("adding. cmd: %s, path: %s \n", promptCmd, path)

	return nil
}

// listCommands use $PATH to find directories
// Then reads each directory and looks for executables
func listCommands() ([]string, error) {
	// Split $PATH directories into slice
	paths := strings.Split(os.Getenv("PATH"), ":")
	var commands []string

	// created buffered error chan
	errc := make(chan error, 1)

	// sync go routines
	var wg sync.WaitGroup

	// find commands appends results to commands slice
	findCommands := func(p string) {
		defer wg.Done()

		files, err := ioutil.ReadDir(p)
		if err != nil {
			errc <- err // write err into error chan
			return
		}

		for _, f := range files {
			m := f.Mode()

			// Check if file is executable
			if m&0111 != 0 {
				commands = append(commands, f.Name())
			}
		}

		errc <- nil // write nil into error chan
	}

	// Check each path for commands
	for _, p := range paths {
		wg.Add(1)
		go findCommands(p)

		// read any error that is in error chan
		if err := <-errc; err != nil {
			return nil, err
		}
	}

	wg.Wait() // Wait for the paths to be checks

	return commands, nil
}

// // createAlias creates .sh file to alias commands
// func createAlias(commandList []string) error {
// 	fmt.Println("create aliases for: ", commandList)
//
// 	return nil
// }

// stats TODO print stats and usage of r
func stats() {
	fmt.Println("stats")
}
