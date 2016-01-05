#r
[![Build Status](https://travis-ci.org/jesselucas/r.svg?branch=master)](https://travis-ci.org/jesselucas/r)

`r` stores all successfully executed commands per directory. Allowing you to easily see what command you last executed. You can also sort by most used commands, recently used and see commands from all directories.

![Terminal Demo](https://dl.dropboxusercontent.com/s/3w7h92ksza871g8/r_terminal.gif)

## Requirements
* OS X / Linux
* Bash

## Installation
### Homebrew:
`brew tap jesselucas/r`

`brew install r`

### Go
* `go get -u github.com/jesselucas/r`
* `r -install` which will add `.r.sh` to home directory and source in `.bashrc`
* or manually add `.r.sh` to your `.bashrc`
  * ex. `. $GOPATH/src/github.com/jesselucas/r/.r.sh`

## Usage
By default `r` shows bash history per directory and is sorted by last used.

You can see all history by using the `-global` flag.

```
Usage of r:
  -install
    installs r.sh to .bashrc
  -global
    show all commands stored by r
  -g
    show all commands stored by r (shorthand)
  -u	sort commands by usage rather than last used (shorthand)
  -usage
    sort commands by usage rather than last used
```
### Example
* Type `r` in any directory and it will prompt `r>`.
* Press `tab` key to see all history.
* Or start typing command and press `tab` to filter history.
* Use `tab` or `arrow` keys to navigate history items.

## Notes
* Set the Directory and Global history in your `.bashrc`
```
# r settings
export R_DIRHISTORY=30 # total to save for directory history
export R_GLOBALHISTORY=100 # total to save for global history
# export R_SORTBYUSAGE=1 # turn this on to default sorting by usage
```

## TODOs
* Write test!
* ~~Make history limit an environment variable~~
* ~~Create flag to see history for all directories~~
* ~~Create flag to sort by most used rather than the default last used.~~
* Create brew formula
* Improve stability of .r.sh
* Make compatible with zsh

## Special Thanks
* [github.com/boltdb/bolt](https://github.com/boltdb/bolt)
* [github.com/chzyer/readline](https://github.com/chzyer/readline)
