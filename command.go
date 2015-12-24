package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

// Sort by last used
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

// Sort by usage
type byUsage []*command

func (s byUsage) Len() int {
	return len(s)
}

func (s byUsage) Less(i, j int) bool {
	return s[i].info.count > s[j].info.count
}

func (s byUsage) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
