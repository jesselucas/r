package r

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Command struct stores the name and CommandInfo for each
// shell command stored in the r database
type Command struct {
	Name string
	Info *CommandInfo
}

// CommandInfo struct is stored as the value to commands
type CommandInfo struct {
	Time  time.Time
	Count int
}

func (ci *CommandInfo) String() string {
	// Store the time in RFC3339 format for easy parsing
	return fmt.Sprintf("%s%s%d", ci.Time.Format(time.RFC3339), ",", ci.Count)
}

// Update method will update the time and count of CommandInfo
func (ci *CommandInfo) Update(ciString string) {
	info := strings.Split(ciString, ",")

	count, err := strconv.Atoi(info[1])
	if err != nil {
		count = 0
	}

	ci.Time = time.Now()
	ci.Count = count + 1
}

// NewFromString creates a new CommandInfo struct from a string
func (ci *CommandInfo) NewFromString(ciString string) *CommandInfo {
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

	ci.Time = date
	ci.Count = count

	return ci
}

// ByTime sorts by last used
type byTime []*Command

// Len used for sorting
func (s byTime) Len() int {
	return len(s)
}

// Less used for sorting
func (s byTime) Less(i, j int) bool {
	return s[i].Info.Time.After(s[j].Info.Time)
}

// Swap used for sorting
func (s byTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ByUsage sorts by usage
type byUsage []*Command

// Len used for sorting
func (s byUsage) Len() int {
	return len(s)
}

// Less used for sorting
func (s byUsage) Less(i, j int) bool {
	return s[i].Info.Count > s[j].Info.Count
}

// Swap used for sorting
func (s byUsage) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
