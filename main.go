package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chzyer/readline"
)

var db *bolt.DB

// commandInfo struct is stored as the value to commands
type commandInfo struct {
	time  time.Time
	count int
}

func (ci commandInfo) String() string {
	// Store the time in RFC3339 format for easy parsing
	return fmt.Sprintf("%s%s%d", ci.time.Format(time.RFC3339), ",", ci.count)
}

func (ci commandInfo) Update(ciString string) commandInfo {
	info := strings.Split(ciString, ",")
	newCI := commandInfo{}

	count, err := strconv.Atoi(info[1])
	if err != nil {
		count = 0
	}

	newCI.time = time.Now()
	newCI.count = count + 1

	return newCI
}

func (ci commandInfo) NewFromString(ciString string) commandInfo {
	info := strings.Split(ciString, ",")
	newCI := commandInfo{}

	// Parse the time as RFC3339 format
	date, err := time.Parse(time.RFC3339, info[0])
	if err != nil {
		date = time.Now()
	}

	count, err := strconv.Atoi(info[1])
	if err != nil {
		count = 0
	}

	newCI.time = date
	newCI.count = count

	return newCI
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

	// Check if `add` flag is passed
	if *addPtr != "" {
		args := strings.Split(*addPtr, ":")
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
	results, err := showResults()
	if err != nil {
		log.Panic(err)
	}

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
		if !containsCmd(line, results) {
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
func showResults() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// results := []string{"git status", "git clone", "go install", "cd /Users/jesse/", "cd /Users/jesse/gocode/src/github.com/jesselucas", "ls -Glah"}
	var results []string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DirectoryBucket"))
		pathBucket := b.Bucket([]byte(wd))
		return pathBucket.ForEach(func(k, v []byte) error {
			// ci := commandInfo{}
			fmt.Printf("%s: %s \n", string(k), string(v))
			// fmt.Printf("%s: %s \n", string(k), ci.NewFromString(string(v)))
			results = append(results, string(k))
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

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

// add checks if command being passed is in the listCommands
// then stores the command and workding directory
func add(path string, promptCmd string) error {
	// get the first command in the promptCmd string
	cmd := strings.Split(promptCmd, " ")[0]

	commands, err := listCommands()
	if err != nil {
		return err
	}

	// check if the command is valid
	if !containsCmd(cmd, commands) {
		return nil
	}

	// Add command to db
	// fmt.Printf("adding. cmd: %s, path: %s \n", promptCmd, path)
	return db.Update(func(tx *bolt.Tx) error {
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

		// Don't store if the command is r
		if cmd == "r" {
			return nil
		}

		// Create commandInfo struct
		ci := commandInfo{}
		ci.time = time.Now()
		ci.count = 1

		// Check if there is a command info value already
		v := cmdBucket.Get([]byte(promptCmd))
		if v != nil {
			// There is a previous command info value
			// Let's update the count and time
			ci = ci.Update(string(v))
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
			ci = ci.Update(string(v))
		}

		err = pathBucket.Put([]byte(promptCmd), []byte(ci.String()))
		if err != nil {
			return err
		}

		return nil
	})
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
