package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/forestgiant/semver"
	"github.com/jesselucas/r"
)

var (
	rSourceName = ".r.sh" // File name of Bash script
)

func main() {
	// Set Semantic Version
	err := semver.SetVersion(r.Version)
	if err != nil {
		log.Fatal(err)
	}

	// Setup flags
	globalUsage := "show all commands stored"
	globalPtr := flag.Bool("global", false, globalUsage)
	flag.BoolVar(globalPtr, "g", false, globalUsage+" (shorthand)")

	// Change sorting flag based on environment variables
	var sortTimePtr *bool
	var sortUsagePtr *bool

	if os.Getenv("R_SORTBYUSAGE") == "1" {
		sortTimeUsage := "sort commands by directory"
		sortTimePtr = flag.Bool("time", false, sortTimeUsage)
		flag.BoolVar(sortTimePtr, "t", false, sortTimeUsage+" (shorthand)")
	} else {
		sortUsageUsage := "sort commands by usage rather than last used"
		sortUsagePtr = flag.Bool("usage", false, sortUsageUsage)
		flag.BoolVar(sortUsagePtr, "u", false, sortUsageUsage+" (shorthand)")
	}

	commandPtr := flag.Bool("command", false, "show last command selected")
	addPtr := flag.String("add", "", "adds command and path to history")
	installPtr := flag.Bool("install", false, fmt.Sprintf("installs %s to .bashrc", rSourceName))
	flag.Parse()

	// Create new r Session
	s := new(r.Session)
	if os.Getenv("R_SORTBYUSAGE") == "1" {
		s.SortTime = *sortTimePtr
	} else {
		s.SortUsage = *sortUsagePtr
	}
	s.Global = *globalPtr

	// Setup bolt db path
	homeDir, err := homeDirectory()
	if err != nil {
		log.Fatal(err)
	}
	s.BoltPath = filepath.Join(homeDir, ".r.db")

	if *commandPtr {
		err = s.PrintLastCommand()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check if `add` flag is passed
	if *addPtr != "" {
		args := strings.Split(*addPtr, "^_")
		if len(args) != 2 {
			fmt.Println("Could not add command.")
			os.Exit(1)
		}

		err := s.Add(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	// Check if .r.sh is installed
	if *installPtr {
		err := install()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// check if the db buckets are empty
	err = s.CheckForHistory()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// reset last command to blank
	// set line as stored command
	err = s.ResetLastCommand()
	if err != nil {
		log.Fatal(err)
	}

	readLine(s)
}

// readLine used the readline library create a prompt to
// show the command history
func readLine(s *r.Session) {
	// Create completer from results
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	var results []*r.Command
	if !s.Global {
		results, err = s.ResultsDirectory(wd)
		if err != nil {
			log.Panic(err)
		}
	} else {
		results, err = s.ResultsGlobal()
		if err != nil {
			log.Panic(err)
		}
	}

	var pcItems []readline.PrefixCompleterInterface
	for _, result := range results {
		pcItems = append(pcItems, readline.PcItem(result.Name))
	}
	var completer = readline.NewPrefixCompleter(pcItems...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "r> ",
		AutoComplete: completer,
	})
	if err != nil {
		log.Panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		line = strings.TrimSpace(line)

		// Only execute if the command typed is in the list of results
		// cmdNames := namesOfCmds(results)
		// if !containsCmd(line, cmdNames) {
		// 	fmt.Println("Command not found in `r` history.")
		// 	os.Exit(0)
		// }

		// The command was found and will be executed so add it to the DB to update
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error executing command.")
			os.Exit(1)
		}
		s.Add(wd, line)

		if err != nil {
			fmt.Println("Error storing command.")
			os.Exit(1)
		}

		// Store last command
		err = s.StoreLastCommand(line)
		if err != nil {
			fmt.Println("Error storing command.")
			os.Exit(1)
		}

		os.Exit(0)
	}
}

// Install will take add .r.sh to you Bash config
func install() error {
	if installed() {
		fmt.Println("r is already installed.")
		return nil
	}

	// install .r.sh
	path, err := bashPath()
	if err == nil {
		err = sourceR(path)
		if err != nil {
			return err
		}

		fmt.Println("r successfully installed! Restart your bash shell.")
		return nil
	}

	return errors.New("Could not install r")
}
