package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

// const dbTestPath = "test.db"

type testDB struct {
	*bolt.DB
	TestPath string
}

func (t *testDB) New() (*testDB, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	t.TestPath = f.Name()

	return t, nil
}

func (t *testDB) Open() error {
	db, err := bolt.Open(t.TestPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	t.DB = db
	if err != nil {
		return err
	}

	return nil
}

func (t *testDB) Close() {
	defer os.Remove(t.TestPath)
	t.DB.Close()
}

func TestResetLastCommand(t *testing.T) {
	db := new(testDB)
	db, err := db.New()
	if err != nil {
		t.Error(err)
	}

	resetLastCommand(db.TestPath)

	err = db.Open()
	if err != nil {
		t.Error(err)
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
		t.Error(err)
	}

	if val != "" {
		t.Error("last command should be blank, but has value:", val)
	}
}

func TestCheckForHistory(t *testing.T) {
	db := new(testDB)
	db, err := db.New()
	if err != nil {
		t.Error(err)
	}

	err = checkForHistory(db.TestPath, false)
	if err.Error() != "r doesn't have a history. Execute commands to build one" {
		t.Error("There shouldn't be a history")
	}

	db.Open()
	// check for global bucket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(globalCommandBucket))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}
	db.DB.Close() // Close boltDB so checkForHistory can open

	err = checkForHistory(db.TestPath, false)
	if err.Error() != "Current directory doesn't have a history. Execute commands to build one" {
		t.Error("There should be a global bucket", err)
	}

	db.Open()
	// check for global bucket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(directoryBucket))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}
	db.DB.Close() // Close boltDB so checkForHistory can open

	err = checkForHistory(db.TestPath, false)
	if err.Error() != "Current directory doesn't have a history. Execute commands to build one" {
		t.Error("There should be a global bucket", err)
	}

	// Close and delete TestDB
	db.Close()

}
