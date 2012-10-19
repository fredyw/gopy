gopy
====

A small utility to list and and copy files written in Go

How to build
------------
`go build gopy.go`

Usage
-----
    ./gopy
      -copy=false: copy operation
      -directory="": directory (for list & copy) - mandatory
      -help=false: help
      -input="": input file (for copy) - mandatory
      -list=false: list operation
      -nodir=false: don't include directories (for list) - optional
      -nofile=false: don't include files (for list) - optional
      -output="": output file (for list) - mandatory
      -recursive=false: recursive (for list) - optional
