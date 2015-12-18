#!/bin/bash

# This will run before any command is executed.
function pre() {
  if [ -z "$R_AT_PROMPT" ]; then
    return
  fi
  unset R_AT_PROMPT

  # Keep reference to what command was executed
  LAST_CMD="${1##*/}"
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

  # Test if LAST_CMD was r then run any command selected
  if [ "$LAST_CMD" == "r" ]; then
    last_r_cmd=$(r --command)
    eval $last_r_cmd

    return
  fi
}

# Run post after command
PROMPT_COMMAND="$PROMPT_COMMAND"$'\n''post'

# function _r_complete() {
#   local cur prev opts args cmd
#   COMPREPLY=()
#   cmd="${COMP_WORDS[0]}" # command
#   args="${COMP_LINE#r }"
#   cur="${COMP_WORDS[COMP_CWORD]}"
#
#   # prev="${COMP_WORDS[COMP_CWORD-1]}" # previous word
#
#   if [ -z "$args" ]; then
#     opts=$( r --complete "$cmd")
#   else
#     opts=$( r --complete $args)
#   fi
#   # opts=$( r --complete $args)
#
#   local IFS=$'\n'
#   COMPREPLY=( $( compgen -W "$opts" -- "$cur" ) )
#
# }
# complete -F _r_complete r
