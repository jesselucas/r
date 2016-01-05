package main

const rBashFile = `#!/bin/bash

# This will run before any command is executed.
pre() {
  if [ -z "$R_AT_PROMPT" ]; then
    return
  fi
  unset R_AT_PROMPT

  # Keep reference to what command was executed
  R_LAST_CMD="${1##*/}"
  export R_PWD
  R_PWD=$(pwd)
  CMD=$*

}

# Set trap to reun pre before command
trap 'pre $BASH_COMMAND' DEBUG

# This will run after the execution of the previous full command line.  We don't
# want post to execute when first starting a bash session (FIRST_PROMPT)
R_FIRST_PROMPT=1
post() {
  R_AT_PROMPT=1

  if [ -n "$R_FIRST_PROMPT" ]; then
    unset R_FIRST_PROMPT
    return
  fi

  local last_code=$?
  # Don't add if the status errored
  if [ "$last_code" -eq 0 ]; then
    # Add current directory and command to r
    r --add "$R_PWD:$CMD"
  fi

  # Test if LAST_CMD was r then run any command selected
  if [ "$R_LAST_CMD" = "r" ]; then
    local last_r_cmd
    last_r_cmd=$(r --command)
    if [ -z "$last_r_cmd" ]; then
      return
    fi

    # execute command
    eval $last_r_cmd

    # save command to bash history
    history -s $last_r_cmd

    return
  fi
}

# Run post after command
PROMPT_COMMAND="${PROMPT_COMMAND}"$'\n'"post;";
`
