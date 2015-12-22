package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chzyer/readline"
)

var db *bolt.DB

type command struct {
	name string
	info *commandInfo
}

// commandInfo struct is stored as the value to commands
type commandInfo struct {
	time  time.Time
	count int
}

func (ci *commandInfo) String() string {
	// Store the time in RFC3339 format for easy parsing
	return fmt.Sprintf("%s%s%d", ci.time.Format(time.RFC3339), ",", ci.count)
}

func (ci *commandInfo) Update(ciString string) {
	info := strings.Split(ciString, ",")

	count, err := strconv.Atoi(info[1])
	if err != nil {
		count = 0
	}

	ci.time = time.Now()
	ci.count = count + 1
}

func (ci *commandInfo) NewFromString(ciString string) *commandInfo {
	info := strings.Split(ciString, ",")

	// Parse the time as RFC3339 format
	date, err := time.Parse(time.RFC3339, info[0])
	if err != nil {
		date = time.Now()
	}

	count, err := strconv.Atoi(info[1])
	if err != nil {
		count = 0
	}

	ci.time = date
	ci.count = count

	return ci
}

type byTime []*command

func (s byTime) Len() int {
	return len(s)
}

func (s byTime) Less(i, j int) bool {
	return s[i].info.time.After(s[j].info.time)
}

func (s byTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func main() {
	// Setup flags
	// statsPtr := flag.Bool("stats", false, "show stats and usage of `r`")
	// completePtr := flag.String("complete", "", "show all results for `r`")
	commandPtr := flag.Bool("command", false, "show last command selected from `r`")
	addPtr := flag.String("add", "", "show stats and usage of `r`")
	flag.Parse()

	// Setup bolt db
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	boltPath := filepath.Join(usr.HomeDir, ".r.db")
	// It will be created if it doesn't exist.
	db, err = bolt.Open(boltPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if `results` flag is passed
	// if *completePtr != "" {
	// 	results := showResults(*completePtr)
	// 	for _, result := range results {
	// 		fmt.Println(result)
	// 	}
	// 	os.Exit(0)
	// }

	if *commandPtr {
		printLastCommand()
		os.Exit(0)
	}

	// TODO add a global flag to see all command history

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

	// reset last command to blank
	// set line as stored command
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("command"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("command"), []byte(""))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	readLine()
}

// readLine used the readline library create a prompt to
// show the command history
func readLine() {
	// create completer from results
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	results, err := results(wd)
	if err != nil {
		log.Panic(err)
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

		cmdNames := namesOfCmds(results)

		// Only execute if the command typed is in the list of results
		if !containsCmd(line, cmdNames) {
			fmt.Println("Command not found in `r` history.")
			os.Exit(0)
		}

		// The command was found and will be executed so add it to the DB to update
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error executing command.")
			os.Exit(1)
		}
		add(wd, line)

		// set line as stored command
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("command"))
			if err != nil {
				return err
			}

			err = b.Put([]byte("command"), []byte(line))
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			fmt.Println("Error storing command.")
			os.Exit(1)
		}

		os.Exit(0)
	}
}

// printLastCommand is used with the --command flag
// it shows the last command selected from the readline prompt
func printLastCommand() {
	var val string
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("command"))
		if err != nil {
			return err
		}

		val = string(b.Get([]byte("command")))
		return nil
	})

	fmt.Println(val)
}

// showResults reads the boltdb and returns the command history
// based on your current working directory
func results(path string) ([]*command, error) {
	// dir := filepath.Dir(path)

	// results := []string{"git status", "git clone", "go install", "cd /Users/jesse/", "cd /Users/jesse/gocode/src/github.com/jesselucas", "ls -Glah"}
	var results []*command
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DirectoryBucket"))
		pathBucket := b.Bucket([]byte(path))
		return pathBucket.ForEach(func(k, v []byte) error {
			cmd := new(command)
			ci := new(commandInfo)
			cmd.name = string(k)
			cmd.info = ci.NewFromString(string(v))

			results = append(results, cmd)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	// Sorty by last command first
	sort.Sort(byTime(results))
	// for _, cmd := range results {
	// 	fmt.Printf("%s: %s \n", cmd.name, cmd.info.time)
	// }

	return results, nil

	// filter
	// var filtered []string
	// for _, result := range results {
	// 	if strings.HasPrefix(result, input) {
	// 		filtered = append(filtered, result)
	// 	}
	// }
	//
	// return filtered

}

func globalResults() ([]*command, error) {
	// Now get all the commands stored
	var results []*command
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("CommandBucket"))
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

	if err != nil {
		return nil, err
	}

	// Sorty by last command first
	sort.Sort(byTime(results))
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

	// Add command to db
	err = db.Update(func(tx *bolt.Tx) error {
		directoryBucket, err := tx.CreateBucketIfNotExists([]byte("DirectoryBucket"))
		if err != nil {
			return err
		}

		pathBucket, err := directoryBucket.CreateBucketIfNotExists([]byte(path))
		if err != nil {
			return err
		}

		// Store path and command for contextual path sorting
		cmdBucket, err := tx.CreateBucketIfNotExists([]byte("CommandBucket"))
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
	// TODO move this to an environment variable
	numberToPruneDir := 20
	numberToPruneGlobal := 100
	pruneGlobal := true
	prunePath := true

	results, err := results(path)
	if err != nil {
		return err
	}

	if len(results) <= numberToPruneDir {
		prunePath = false
	}

	// List the global commands
	globalResults, err := globalResults()
	if err != nil {
		return err
	}

	// set pruneGlobal to true if there isn't enough
	if len(globalResults) <= numberToPruneGlobal {
		pruneGlobal = false
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if prunePath {
			directoryBucket, err := tx.CreateBucketIfNotExists([]byte("DirectoryBucket"))
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
			cmdBucket, err := tx.CreateBucketIfNotExists([]byte("CommandBucket"))
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
