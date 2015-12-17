#!/bin/bash

# This will run before any command is executed.
function pre() {
  if [ -z "$R_AT_PROMPT" ]; then
    return
  fi
  unset R_AT_PROMPT

  # Keep reference to what command was executed
  # local cmd="${1##*/}"
  local cmd=$*

  # Add current directory and command to `r`
  r --add "$(\pwd):$cmd"
}

# Set trap to reun pre before command
trap 'pre $BASH_COMMAND' DEBUG

# This will run after the execution of the previous full command line.  We don't
# want `post` to execute when first starting a bash session (FIRST_PROMPT)
R_FIRST_PROMPT=1
function post() {
  R_AT_PROMPT=1

  if [ -n "$R_FIRST_PROMPT" ]; then
    unset R_FIRST_PROMPT
    return
  fi

  # Post command logic
}

# Run post after command
PROMPT_COMMAND="$PROMPT_COMMAND"$'\n''post'
