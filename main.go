package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chzyer/readline"
	"github.com/forestgiant/semver"
)

var (
	boltPath     string
	sortUsagePtr *bool
	sortTimePtr  *bool
)

const (
	globalCommandBucket = "GlobalCommandBucket"
	directoryBucket     = "DirectoryBucket"
	lastCommandBucket   = "lastCommandBucket"
	rSourceName         = ".r.sh"
)

func main() {
	// Set Semantic Version
	err := semver.SetVersion("0.3.2")
	if err != nil {
		log.Fatal(err)
	}

	// Setup flags
	globalUsage := "show all commands stored"
	globalPtr := flag.Bool("global", false, globalUsage)
	flag.BoolVar(globalPtr, "g", false, globalUsage+" (shorthand)")

	// Change sorting flag based on environment variables
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

	// Setup bolt db path
	homeDir, err := homeDirectory()
	if err != nil {
		log.Fatal(err)
	}
	boltPath = filepath.Join(homeDir, ".r.db")

	if *commandPtr {
		err = printLastCommand(boltPath)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check if `add` flag is passed
	if *addPtr != "" {
		args := strings.Split(*addPtr, ":")
		if len(args) != 2 {
			fmt.Println("Could not add command.")
			os.Exit(1)
		}

		err := add(args[0], args[1])
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

		fmt.Println("r successfully installed! Restart your bash shell.")
		os.Exit(0)
	}

	// check if the db buckets are empty
	err = checkForHistory(boltPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// reset last command to blank
	// set line as stored command
	err = resetLastCommand(boltPath)
	if err != nil {
		log.Fatal(err)
	}

	readLine(*globalPtr)
}

func install() error {
	if installed() {
		return errors.New("r is already installed")
	}

	// install .r.sh
	path, err := bashPath()
	if err == nil {
		err = sourceR(path)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("Could not install r")
}

func resetLastCommand(boltPath string) error {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error resetLastCommand")
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(lastCommandBucket))
		if err != nil {
			return err
		}

		err = b.Put([]byte("command"), []byte(""))
		if err != nil {
			return err
		}
		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	return nil
}

func checkForHistory(boltPath string) error {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error checkForHistory")
		return err
	}

	// Check if global bucket is empty. if it is return
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(globalCommandBucket))
		if b == nil {
			return errors.New("r doesn't have a history. Execute commands to build one")
		}

		// Check if current wording directy has a history
		// if it doesn't return
		wd, err := os.Getwd()
		if err != nil {
			return errors.New("Current directory doesn't have a history. Execute commands to build one")
		}

		b = tx.Bucket([]byte(directoryBucket))
		if b == nil {
			return errors.New("Current directory doesn't have a history. Execute commands to build one")
		}

		pathBucket := b.Bucket([]byte(wd))
		if pathBucket == nil {
			return errors.New("Current directory doesn't have a history. Execute commands to build one")
		}

		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	return nil
}

// readLine used the readline library create a prompt to
// show the command history
func readLine(global bool) {
	// Create completer from results
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	var results []*command
	if !global {
		results, err = resultsDirectory(boltPath, wd)
		if err != nil {
			log.Panic(err)
		}
	} else {
		results, err = resultsGlobal(boltPath)
		if err != nil {
			log.Panic(err)
		}
	}

	var pcItems []*readline.PrefixCompleter
	for _, result := range results {
		pcItems = append(pcItems, readline.PcItem(result.name))
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
		add(wd, line)

		if err != nil {
			fmt.Println("Error storing command.")
			os.Exit(1)
		}

		// Store last command
		err = storeLastCommand(boltPath, line)
		if err != nil {
			fmt.Println("Error storing command.")
			os.Exit(1)
		}

		os.Exit(0)
	}
}

func storeLastCommand(boltPath string, line string) error {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error storeLastCommand")
		return err
	}

	// Set line as stored command
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(lastCommandBucket))
		if err != nil {
			return err
		}

		err = b.Put([]byte("command"), []byte(line))
		if err != nil {
			return err
		}
		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	return nil
}

// printLastCommand is used with the --command flag
// it shows the last command selected from the readline prompt
func printLastCommand(boltPath string) error {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error printLastCommand")
		return err
	}

	var val string
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(lastCommandBucket))
		if err != nil {
			return err
		}

		val = string(b.Get([]byte("command")))
		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	fmt.Println(val)
	return nil
}

func sortCommands(results []*command) {
	// Check for environment variable for usage sorting
	if os.Getenv("R_SORTBYUSAGE") == "1" {
		if !*sortTimePtr {
			sort.Sort(byUsage(results))
		} else {
			sort.Sort(byTime(results))
		}
		return
	}

	// Check for usage flag
	if !*sortUsagePtr {
		sort.Sort(byTime(results))
	} else {
		sort.Sort(byUsage(results))
	}
}

// showResults reads the boltdb and returns the command history
// based on your current working directory
func resultsDirectory(boltPath string, path string) ([]*command, error) {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error results")
		return nil, err
	}

	var results []*command
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(directoryBucket))
		pathBucket := b.Bucket([]byte(path))
		return pathBucket.ForEach(func(k, v []byte) error {
			cmd := new(command)
			ci := new(commandInfo)
			cmd.name = string(k)
			cmd.info = ci.NewFromString(string(v))

			if cmd.name == `[ "$LAST_CMD" = "r" ]` {
				return nil
			}

			results = append(results, cmd)
			return nil
		})
	})

	db.Close()

	if err != nil {
		return nil, err
	}

	// Sort commands
	sortCommands(results)

	// Print results (Used for testing)
	// for _, cmd := range results {
	// 	fmt.Printf("%s: %s \n", cmd.name, cmd.info.count)
	// }

	return results, nil
}

func resultsGlobal(boltPath string) ([]*command, error) {
	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error globalResults")
		return nil, err
	}

	// Now get all the commands stored
	var results []*command
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(globalCommandBucket))
		err := b.ForEach(func(k, v []byte) error {
			command := new(command)
			ci := new(commandInfo)
			command.name = string(k)
			command.info = ci.NewFromString(string(v))
			results = append(results, command)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	db.Close()

	if err != nil {
		return nil, err
	}

	// Sort commands
	sortCommands(results)

	// Print results (Used for testing)
	// for _, cmd := range results {
	// 	fmt.Printf("%s: %s \n", cmd.name, cmd.info.time)
	// }

	return results, nil
}

// add checks if command being passed is in the listCommands
// then stores the command and workding directory
func add(path string, promptCmd string) error {
	// get the first command in the promptCmd string
	cmd := strings.Split(promptCmd, " ")[0]

	// Don't store if the command is r
	if cmd == "r" {
		return nil
	}

	commands, err := listCommands()
	if err != nil {
		return err
	}

	// check if the command is valid
	if !containsCmd(cmd, commands) {
		return nil
	}

	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error add")
		return err
	}

	// Add command to db
	err = db.Update(func(tx *bolt.Tx) error {
		dBucket, err := tx.CreateBucketIfNotExists([]byte(directoryBucket))
		if err != nil {
			return err
		}

		pathBucket, err := dBucket.CreateBucketIfNotExists([]byte(path))
		if err != nil {
			return err
		}

		// Store path and command for contextual path sorting
		cmdBucket, err := tx.CreateBucketIfNotExists([]byte(globalCommandBucket))
		if err != nil {
			return err
		}

		// Create commandInfo struct
		ci := new(commandInfo)
		ci.time = time.Now()
		ci.count = 1

		// Check if there is a command info value already
		v := cmdBucket.Get([]byte(promptCmd))
		if v != nil {
			// There is a previous command info value
			// Let's update the count and time
			ci.Update(string(v))
		}

		err = cmdBucket.Put([]byte(promptCmd), []byte(ci.String()))
		if err != nil {
			return err
		}

		// Now let's do the same thing for the pathBucket
		v = pathBucket.Get([]byte(promptCmd))
		if v != nil {
			// There is a previous command info value
			// Let's update the count and time
			ci.Update(string(v))
		}

		err = pathBucket.Put([]byte(promptCmd), []byte(ci.String()))
		if err != nil {
			return err
		}

		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	// now prune the older commands
	err = prune(path)
	if err != nil {
		return err
	}

	return nil
}

// prune deletes commands from a directory bucket and overall bucket
func prune(path string) error {
	// Set number to prune from envar
	numberToPruneDir, err := strconv.Atoi(os.Getenv("R_DIRHISTORY"))
	if err != nil {
		numberToPruneDir = 30
	}

	numberToPruneGlobal, err := strconv.Atoi(os.Getenv("R_GLOBALHISTORY"))
	if err != nil {
		numberToPruneGlobal = 100
	}

	pruneGlobal := true
	prunePath := true

	results, err := resultsDirectory(boltPath, path)
	if err != nil {
		return err
	}

	if len(results) <= numberToPruneDir {
		prunePath = false
	}

	// List the global commands
	globalResults, err := resultsGlobal(boltPath)
	if err != nil {
		return err
	}

	// set pruneGlobal to true if there isn't enough
	if len(globalResults) <= numberToPruneGlobal {
		pruneGlobal = false
	}

	db, err := bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error prune")
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if prunePath {
			directoryBucket, err := tx.CreateBucketIfNotExists([]byte(directoryBucket))
			if err != nil {
				return err
			}

			pathBucket, err := directoryBucket.CreateBucketIfNotExists([]byte(path))
			if err != nil {
				return err
			}

			pruneDirResults := results[numberToPruneDir:]
			for _, cmd := range pruneDirResults {
				pathBucket.Delete([]byte(cmd.name))
			}
		}
		// Prune stored global commands
		if pruneGlobal {
			// Store path and command for contextual path sorting
			cmdBucket, err := tx.CreateBucketIfNotExists([]byte(globalCommandBucket))
			if err != nil {
				return err
			}

			pruneGlobalResults := globalResults[numberToPruneGlobal:]
			for _, cmd := range pruneGlobalResults {
				cmdBucket.Delete([]byte(cmd.name))
			}
		}

		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	return nil
}

// namesOfCmds takes a slice of command structs and return
// a slice with just their names
func namesOfCmds(cmds []*command) []string {
	var names []string
	for _, cmd := range cmds {
		names = append(names, cmd.name)
	}

	return names
}

// containsCmd checks if a command string is is in a slice of strings
func containsCmd(cmd string, commands []string) bool {
	for _, c := range commands {
		// check first command against list of commands
		if c == cmd {
			return true
		}
	}
	return false
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
