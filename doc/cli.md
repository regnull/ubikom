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

## Create a Private Key

Before you can do pretty much anything, you need to create a private key. You can do it by using "create key" command:

```
$ ubikom-cli create key --out=secret.key
Passphrase (enter for none):
Confirm passphrase (enter for none):
15:43:48 WRN saving private key without passphrase
15:43:48 INF private key saved location=secret.key
```

It is recommended that you use a passphrase when you create a key. If you don't, anyone who can access this file
will be able to impersonate you.

## Getting Key Information

Now that you have your key, you can get various details about it.

### Get Key Address

To get the Bitcoin-style key address, use "get address" command:

```
$ ubikom-cli get address --key=secret.key
1GCFeppSWHPFwvFdAzU7N7CLA6A9jFVX3J
```

If your key was saved with a passphrase, you will be prompted for one.

### Get Public Key

To get the public key associated with this private key, use "get public-key" command:

```
$ ubikom-cli get public-key --key=secret.key
y55y9N5aRJ3wbvV2oULsJhECrWE26be5LHHV4iWcrToE
```

### Get Key Mnemonic

Key mnemonic allows you to restore the key (using "create key --from-mnemonic" command) if you don't have the key file.
Use this one with caution! You want to do it on a secure machine (ideally air-gapped). Write the mnemonic down
and keep it in a safe place.

```
$ ubikom-cli get mnemonic --key=secret.key
1: 	glass
2: 	vessel
3: 	transfer
4: 	broken
5:  ....
```

## Registering keys, names, and addresses

### Registering your public key

Before using your private key, you must register it. When a key is registered, its information
is stored in a public registry where it can be accessed by other users. For example, some
other user might want to access your public key to encrypt mail addressed to you. Having public
registry solves other problems as well. If your key is compromised, you can permanently 
disable it in the public registry, rendering it useless.

To register the key, execute the following command:

```
$ ubikom-cli register key --key=secret.key
8:50:38 DBG generating POW...
18:50:38 DBG POW found pow=5f5bf752ad129813
18:50:38 INF key registered successfully
```

When you register a key, you must generate some minimal proof-of-work (POW). This is done to 
reduce name squatting and spamming. Normally, generating POW will only take a few seconds.
