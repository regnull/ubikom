# Using Ubikom CLI

## Prerequisites

You must have Go installed. If you don't, follow the instructions [here](https://golang.org/doc/install).

## Install CLI

The easiest way to install Ubikom CLI is by using go get command:

```
go get github.com/regnull/ubikom/cmd/ubikom-cli
```

The binary will appear under your $GOROOT/bin directory, by default it will be $HOME/go/bin.
Either go to that directory or add it to path. Run ubikom-cli to see the list of available commands:

```
$ ubikom-cli
ubikom-cli allows you to run local and remote Ubikom commands

Usage:
  ubikom-cli [flags]
  ubikom-cli [command]

Available Commands:
  create      Create various things
  disable     Disable something
  get         Get various things
  help        Help about any command
  lookup      Look stuff up
  receive     Receive stuff
  register    Register various things
  send        Send stuff

Flags:
  -h, --help   help for ubikom-cli

Use "ubikom-cli [command] --help" for more information about a command.
```