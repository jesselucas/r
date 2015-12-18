package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/chzyer/readline"
)

func main() {
	// Setup flags
	statsPtr := flag.Bool("stats", false, "show stats and usage of `r`")
	completePtr := flag.String("complete", "", "show all results for `r`")
	addPtr := flag.String("add", "", "show stats and usage of `r`")
	flag.Parse()

	// Check if `stats` flag is passed
	if *statsPtr {
		stats()
		os.Exit(0)
	}

	// Check if `results` flag is passed
	if *completePtr != "" {
		results := showResults(*completePtr)
		for _, result := range results {
			fmt.Println(result)
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

		os.Exit(0)
	}

	readLine()
}

func readLine() {
	// create completer from results
	results := showResults("r")
	var pcItems []*readline.PrefixCompleter
	for _, result := range results {
		pcItems = append(pcItems, readline.PcItem(result))
	}
	var completer = readline.NewPrefixCompleter(pcItems...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		println(line)
	}
}

// setupDB verifies and creates boltDB in ~ home folder
func setupDB() error {
	// TODO setup boltdb
	fmt.Println("setup boltdb")

	return nil
}

func showResults(input string) []string {
	results := []string{"git status", "git clone", "go install", "cd ~", "cd $GOPATH/src/github.com/jesselucas", "ls -la"}

	if input == "r" {
		return results
	}

	// filter
	fmt.Println("filtered: ", input)
	var filtered []string
	for _, result := range results {
		if strings.HasPrefix(result, input) {
			filtered = append(filtered, result)
		}
	}

	return filtered

}

// add checks if command being passed is in the listCommands
// then stores the command and workding directory
func add(path string, promptCmd string) error {
	// get the first command in the promptCmd string
	cmd := strings.Split(promptCmd, " ")[0]

	commands, err := listCommands()
	if err != nil {
		return err
	}

	containsCmd := func() bool {
		for _, c := range commands {
			// check first command against list of commands
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

	wg.Wait() // Wait for the paths to be checked

	return commands, nil
}

// stats TODO print stats and usage of r
func stats() {
	fmt.Println("stats")
}
