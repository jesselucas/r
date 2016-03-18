package r

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const (
	globalCommandBucket = "GlobalCommandBucket" // BoltDB bucket storing all commands
	directoryBucket     = "DirectoryBucket"     // BoltDB bucket storying commands per directory
	lastCommandBucket   = "lastCommandBucket"   // BoltDB bucket storing the last command r selected

	// Version is semantic version for package r and cmd/r
	Version = "0.4.2"
)

// Session is created every time r cmd is ran
type Session struct {
	// Path to store and reference boltdb
	BoltPath string
	// Used to store the bool value from the r cmd global flag
	Global bool
	// SortUsagePtr used to check if the usage flag was used
	SortUsage bool
	// SortTimePtr used to check if the time flag was used
	SortTime bool
}

// ResetLastCommand clears the value in the lastCommandBucket
func (s *Session) ResetLastCommand() error {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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

// CheckForHistory makes sure a directory has history or if the global bool is true
// it will make sure the global bucket has a history
func (s *Session) CheckForHistory() error {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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
		if !s.Global {
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
		}

		return nil
	})

	db.Close()

	if err != nil {
		return err
	}

	return nil
}

// StoreLastCommand takes the line string and stores it
func (s *Session) StoreLastCommand(line string) error {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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

// PrintLastCommand is used with the r cli --command flag
// it shows the last command selected from the readline prompt
func (s *Session) PrintLastCommand() error {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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

func (s *Session) sortCommands(results []*Command) {
	// Check for environment variable for usage sorting
	if os.Getenv("R_SORTBYUSAGE") == "1" {
		if !s.SortTime {
			sort.Sort(byUsage(results))
		} else {
			sort.Sort(byTime(results))
		}
		return
	}

	// Check for usage flag
	if !s.SortUsage {
		sort.Sort(byTime(results))
	} else {
		sort.Sort(byUsage(results))
	}
}

// ResultsDirectory reads the boltdb and returns the command history
// based on your current working directory
func (s *Session) ResultsDirectory(path string) ([]*Command, error) {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error results")
		return nil, err
	}

	var results []*Command
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(directoryBucket))
		pathBucket := b.Bucket([]byte(path))
		return pathBucket.ForEach(func(k, v []byte) error {
			cmd := new(Command)
			ci := new(CommandInfo)
			cmd.Name = string(k)
			cmd.Info = ci.NewFromString(string(v))

			if cmd.Name == `[ "$LAST_CMD" = "r" ]` {
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
	s.sortCommands(results)

	// Print results (Used for testing)
	// for _, cmd := range results {
	// 	fmt.Printf("%s: %s \n", cmd.name, cmd.info.count)
	// }

	return results, nil
}

// ResultsGlobal returns all the results for the global commands bucket
func (s *Session) ResultsGlobal() ([]*Command, error) {
	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("error globalResults")
		return nil, err
	}

	// Now get all the commands stored
	var results []*Command
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(globalCommandBucket))
		err := b.ForEach(func(k, v []byte) error {
			command := new(Command)
			ci := new(CommandInfo)
			command.Name = string(k)
			command.Info = ci.NewFromString(string(v))
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
	s.sortCommands(results)

	// Print results (Used for testing)
	// for _, cmd := range results {
	// 	fmt.Printf("%s: %s \n", cmd.name, cmd.info.time)
	// }

	return results, nil
}

// Add checks if command being passed is in the listCommands
// then stores the command and workding directory
func (s *Session) Add(path string, promptCmd string) error {
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

	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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
		ci := new(CommandInfo)
		ci.Time = time.Now()
		ci.Count = 1

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
	err = s.Prune(path)
	if err != nil {
		return err
	}

	return nil
}

// Prune deletes commands from a directory bucket and overall bucket
func (s *Session) Prune(path string) error {
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

	results, err := s.ResultsDirectory(path)
	if err != nil {
		return err
	}

	if len(results) <= numberToPruneDir {
		prunePath = false
	}

	// List the global commands
	globalResults, err := s.ResultsGlobal()
	if err != nil {
		return err
	}

	// set pruneGlobal to true if there isn't enough
	if len(globalResults) <= numberToPruneGlobal {
		pruneGlobal = false
	}

	db, err := bolt.Open(s.BoltPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
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
				pathBucket.Delete([]byte(cmd.Name))
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
				cmdBucket.Delete([]byte(cmd.Name))
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
func namesOfCmds(cmds []*Command) []string {
	var names []string
	for _, cmd := range cmds {
		names = append(names, cmd.Name)
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
	defer close(errc)

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
