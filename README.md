#r
[![Build Status](https://travis-ci.org/jesselucas/r.svg?branch=master)](https://travis-ci.org/jesselucas/r)

A contextual, path based, bash history. Like `ctrl-r` but for each directory.

## Requirements
* OS X / Linux
* Bash
* Must add `r.sh` to your `.bashrc`

## Installation
* `go get -u github.com/jesselucas/r`
* add `r.sh` to your `.bashrc`
  * ex. `. $GOPATH/src/github.com/jesselucas/r/r.sh`

## Usage
By default `r` shows bash history per directory and is sorted by last used.

You can see all history by using the `-global` flag.

```
Usage of r:
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
* Improve stability of r.sh
* Make compatible with zsh

## Special Thanks
* [github.com/chzyer/readline](https://github.com/chzyer/readline)
