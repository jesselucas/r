#r
[![Build Status](https://travis-ci.org/jesselucas/r.svg?branch=master)](https://travis-ci.org/jesselucas/r)

A contextual, path based, bash history. Like `ctrl-r` but for each directory.

## Requirements
`r` currently only works with bash and you must add `r.sh` to your `.bashrc`

## Installation
* `go get -u github.com/jesselucas/r`
* add `r.sh` to your `.bashrc`
  * ex. `. $GOPATH/src/github.com/jesselucas/r/r.sh`

## Usage
* Type `r` in any directory and it will prompt `r>`.
* Press `tab` key to see all history.
* Or start typing command and press `tab` to filter history.
* Use `tab` or `arrow` keys to navigate history items.

## Notes
* Set the Directory and Global history in your `.bashrc`
```
# r directory history
export R_DIRHISTORY=30

# r global history
export R_GLOBALHISTORY=100
```

## TODOs
* Write test!
* Make history limit an environment variable
* Create flag to see history for all directories
* Improve stability of r.sh
* Make compatible with zsh

## Special Thanks
* [github.com/chzyer/readline](https://github.com/chzyer/readline)
