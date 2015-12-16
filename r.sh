#!/bin/bash

_r() {
  echo "test"

  # keep reference to what command was executed
  local cmd="${1##*/}"

  # Add command and path to r
  r --add "$(\pwd)":$cmd

  # Execute command
  "$cmd"
}

# Look for existing aliases
array=($(alias))
echo "${array[@]}"

# List of commands to alias list is returned from `r --commands`

# Loop through list of commands, and check against existing alias

# alias all commands
alias ls='_r ls'
alias pwd='_r pwd'
