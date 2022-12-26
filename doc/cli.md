# Using Ubikom CLI

  * [Prerequisites](#prerequisites)
  * [Install CLI](#install-cli)
  * [Output Format](#output-format)
  * [Create a Private Key](#create-a-private-key)
    + [Create Key Using Password](#create-key-using-password)
    + [Create Key from Mnemonic](#create-key-from-mnemonic)
  * [Getting Key Information](#getting-key-information)
    + [Get Key Address](#get-key-address)
    + [Get Public Key](#get-public-key)
    + [Get Key Mnemonic](#get-key-mnemonic)
  * [Getting Balance, Funding](#getting-balance--funding)
  * [Registering Names, Updating Configuration](#registering-names--updating-configuration)
    + [Registering Name](#registering-name)
    + [Registering Messaging Endpoint](#registering-messaging-endpoint)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

## Prerequisites

You must have Go installed. If you don't, follow the instructions [here](https://golang.org/doc/install).

## Install CLI

The easiest way to install Ubikom CLI is by using go get command:

```
$ go get github.com/regnull/ubikom/cmd/ubikom-cli
```

Or, if you've already cloned ubikom repo, do:

```
$ cd ubikom-directory/cmd/ubikom-cli
$ go install
```

The binary will appear under your $GOROOT/bin directory, by default it will be $HOME/go/bin.
Either go to that directory or add it to path. Run ubikom-cli to see the list of available commands:

```
$ ubikom-cli allows you to run local and remote Ubikom commands

Usage:
  ubikom-cli [flags]
  ubikom-cli [command]

Available Commands:
  bc          Blockchain-related commands
  completion  Generate the autocompletion script for the specified shell
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

## Output Format

When you request data using ubikom-cli, the output format will be either a string or JSON:

* If the data includes a single value (for example, your account balance), it's printed out as a string;
* If the data includes multiple values (for example, the result of blockchain interaction), it's printed out as JSON.

## Create a Private Key

Before you can do pretty much anything, you need to create a private key. You can do it by using "create key" command:

```
$ ubikom-cli create key --out=secret.key
Passphrase (enter for none):
Confirm passphrase (enter for none):
15:43:48 WRN saving private key without passphrase
15:43:48 INF private key saved location=secret.key
```

The key will be saved as secret.key file in the current directory.

It is recommended that you use a passphrase when you create a key. If you don't, anyone who can access this file
will be able to impersonate you.

### Create Key Using Password

It is possible to construct a private key given two pieces of data: password and salt. This makes it possible,
for example, to transmit a private key over as a "user name" and "password", where salt plays the role of
"user name". Email clients use it to send the email key over to Ubikom proxy using POP3 or SMTP protocol.

```
$ ubikom-cli create key --from-password=supersecretpassword123 --salt=123456 \
  --out=secret1.key
Passphrase (enter for none):
Confirm passphrase (enter for none):
14:37:01 WRN saving private key without passphrase
14:37:01 INF private key saved location=secret1.key
```

If you later use this command with --salt flag and specify the same salt, you will end up with the same key.

Let's make sure this is the case. Generate another key using the same password and salt:

```
$ ubikom-cli create key --from-password=supersecretpassword123 --salt=123456 \
  --out=secret2.key
```

Now we can compute SHA256 hash of both files and compare the hashes:

```
$ sha256sum secret1.key
97a9a2a789a9905d43d8e6922fe2cfc14a05e2aa370b5408291b996e86fa3fa5  secret1.key

$ sha256sum secret2.key
97a9a2a789a9905d43d8e6922fe2cfc14a05e2aa370b5408291b996e86fa3fa5  secret2.key
```

### Create Key from Mnemonic

Another way to create a private key is from mnemonic. A popular way to store a private key is to use a list
of 24 words, as specified in [BIP 39](https://en.bitcoin.it/wiki/BIP_0039).

To re-create private key from mnemonic, use --from-mnemonic flag:

```
$ ubikom-cli create key --from-mnemonic --out=secret.key
```

You will need to enter 24 words one by one, so that the key can be created.

## Getting Key Information

Now that you have your key, you can get various details about it.

### Get Key Address

To get the Bitcoin-style key address, use "get address" command:

```
$ ubikom-cli get address --key=secret.key
1GCFeppSWHPFwvFdAzU7N7CLA6A9jFVX3J
```

If your key was saved with a passphrase, you will be prompted for one.

Most of the time we will be using key's Ethereum-style address, which you get get like so:

```
$ ubikom-cli get ethereum-address --key secret.key
0x27A5f262Be45D99068C157c5A10430ddA252B1f6
```

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

## Getting Balance, Funding

Now that we have our key registered, we can query our balance:

```
$ ADDRESS=$(ubikom-cli get ethereum-address --key secret.key)
$ ubikom-cli bc get balance $ADDRESS --mode=test
15:11:26 WRN using Sepolia testnet
Balance: 0
```

In the first line, we got our Ethereum address and assigned it to the ADDRESS variable. In the second line,
we queried the balance of this address, which, unsurprisingly, is zero. Notice that we are using Sepolia testnet
(--mode=test argument). Without it, we would go to the Ethereum mainnet. The warning message was printed
out to alert us to the fact that we are working on the testnet.

Let's get some funds into our test account. To do that, find a Sepolia faucet and request funds using our 
Ethereum address, which, again, you can obtain using this command:

```
$ ubikom-cli get ethereum-address --key secret.key
```

At the moment when this was written, Sepolia faucets were somewhat unreliable, but if you try a few of them
you can find one that works. Google for 'free Sepolia faucet'. As always, be carefull by not giving your
private information to anyone.

I had some luck using this one: https://sepolia-faucet.com/, which mines Ether in your browser for a few minutes
and then sends the reward to you. Use at your own risk.

After you get some funds, your balance will change:

```
$ ubikom-cli bc get balance $ADDRESS --mode=test
15:29:33 WRN using Sepolia testnet
Balance: 32375000000000000
```

Notice that your balance is in wei, the smallest unit in Ethereum ecosystem. To convert it into Ether, you 
may use a tool like this one: https://eth-converter.com/

## Registering Names, Updating Configuration

### Registering Name

Once we have some funds in our account, we can register a name. After you register a name, you own it - you 
can transfer the ownership to anyone else (if you so choose), or sell it in the future. Here, we will
register a name on Sepolia testnet - the Ethereum mainnet works the same way, but it might not be a great
idea to use ubikom-cli on the mainnet - our keys are for testing only, we don't have industrial strength 
protection that is offered by widely used software and hardware wallets.

Before we register a name, we should create an encryption key. This is optional - if you don't specify
an encryption key, you won't be able to send and receive secure messages. You can always change the
encryption key later (but you will have to pay the gas fees).

The encryption key is just another key. Let's create it:

```
$ ubikom-cli create key --out=encrypt.key
```

Now we can register a new name:

```
$ ubikom-cli bc register name test666 --key=secret.key --enc-key=encrypt.key --mode=test
17:03:20 WRN using Sepolia testnet
17:03:20 DBG using node node-url=https://sepolia.infura.io/v3/8f540714acb24862a8c9a5c3d8568f23
17:03:20 WRN using Sepolia testnet
17:03:20 DBG using contract contract-address=0xcc8650c9cd8d99b62375c22f270a803e7abf0de9
17:03:20 DBG got nonce nonce=1
17:03:20 DBG got gas price gas-price=500000000007
17:03:20 DBG got chain ID chain-id=11155111
tx sent: 0x8375b513f86a5dbcf8342233048af31e4611b93ca01d3926e17ddf7ff0bdfd24
{
  "root": "0x",
  "status": "0x1",
  "cumulativeGasUsed": "0x227a8",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000200000000000000000000000000000010000000000000000000000000000000000000020000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "logs": [
    {
      "address": "0xcc8650c9cd8d99b62375c22f270a803e7abf0de9",
      "topics": [
        "0x1c6eac0e720ec22bb0653aec9c19985633a4fb07971cf973096c2f8e3c37c17f"
      ],
      "data": "0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000027a5f262be45d99068c157c5a10430dda252b1f600000000000000000000000000000000000000000000000000000000000000077465737436363600000000000000000000000000000000000000000000000000",
      "blockNumber": "0x26f669",
      "transactionHash": "0x8375b513f86a5dbcf8342233048af31e4611b93ca01d3926e17ddf7ff0bdfd24",
      "transactionIndex": "0x1",
      "blockHash": "0xabc049143e555d2021f7e97e785f8b71312bb255389ab656043a99fa0351bca3",
      "logIndex": "0x0",
      "removed": false
    }
  ],
  "transactionHash": "0x8375b513f86a5dbcf8342233048af31e4611b93ca01d3926e17ddf7ff0bdfd24",
  "contractAddress": "0x0000000000000000000000000000000000000000",
  "gasUsed": "0x1d5a0",
  "blockHash": "0xabc049143e555d2021f7e97e785f8b71312bb255389ab656043a99fa0351bca3",
  "blockNumber": "0x26f669",
  "transactionIndex": "0x1"
}
```

If you try to register the same name again, you will get an error:

```
$ ubikom-cli bc register name test666 --key=secret.key --enc-key=encrypt.key --mode=test
17:05:08 WRN using Sepolia testnet
17:05:08 DBG using node node-url=https://sepolia.infura.io/v3/8f540714acb24862a8c9a5c3d8568f23
17:05:08 WRN using Sepolia testnet
17:05:08 DBG using contract contract-address=0xcc8650c9cd8d99b62375c22f270a803e7abf0de9
17:05:09 DBG got nonce nonce=2
17:05:09 DBG got gas price gas-price=500000000007
17:05:09 DBG got chain ID chain-id=11155111
tx sent: 0x203a0890e7c588df52fa954d0e0d0a76486b96ba7aefdb2a0887f4fb2c74f4f2
{
  "root": "0x",
  "status": "0x0",
  "cumulativeGasUsed": "0x3cb27",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "logs": [],
  "transactionHash": "0x203a0890e7c588df52fa954d0e0d0a76486b96ba7aefdb2a0887f4fb2c74f4f2",
  "contractAddress": "0x0000000000000000000000000000000000000000",
  "gasUsed": "0x65e2",
  "blockHash": "0x9accdee8cc26b87061988a8896e7f623fef51ee588df7cd62b18d58e5d09ae9e",
  "blockNumber": "0x26f671",
  "transactionIndex": "0x4"
}
17:05:25 ERR transaction failed
17:05:25 FTL failed to register name error="transaction failed"
```

Let's verify that the name was successfully registered:

```
$ ubikom-cli bc lookup name test666 --mode=test
17:10:21 WRN using Sepolia testnet
17:10:21 DBG using node node-url=https://sepolia.infura.io/v3/8f540714acb24862a8c9a5c3d8568f23
17:10:21 WRN using Sepolia testnet
17:10:21 DBG using contract contract-address=0xcc8650c9cd8d99b62375c22f270a803e7abf0de9
{
  "PublicKey": "0x0367714ab1510079fb9c79128f0d3358742dfeaf994bb15190a72715193c8710c3",
  "Owner": "0x27a5f262be45d99068c157c5a10430dda252b1f6",
  "Price": 0
}
```

Here's what we see:
* The public key is our encryption public key (you can get it by running "ubikom-cli get public-key --key=encrypt.key");
* THe owner is us (this is our Ethereum address);
* The price of the name is zero, which means it's not for sale.

### Registering Messaging Endpoint

Before you can start receiving messages, you must register a messaging endpoint. This is the endpoint 
other users will use to send messages to you. Each name may have a number of configuration entries
associated with it, each entry is a key/value pair (both are strings). For messaging, the name of the
configuration entry must be "dms-endpoint". Ubikom provides a public server located at alpha.ubikom.cc:8826,
which can be used as a messaging endpoint. It's fine to use another address (as long as it's running a 
server implementing the messaging protocol). You can run your own server, if you want.

To register a messaging endpoint, run the following command:

```
ubikom-cli bc update config test666 --config-name=dms-endpoint --config-value=alpha.ubikom.cc:8826 --mode=test --key=secret.key --gas-
price=3000000
18:49:14 WRN using Sepolia testnet
18:49:14 DBG using node node-url=https://sepolia.infura.io/v3/8f540714acb24862a8c9a5c3d8568f23
18:49:14 WRN using Sepolia testnet
18:49:14 DBG using contract contract-address=0xcc8650c9cd8d99b62375c22f270a803e7abf0de9
18:49:14 DBG got nonce nonce=3
18:49:14 DBG got gas price gas-price=3000000
18:49:14 DBG got chain ID chain-id=11155111
tx sent: 0x2e5089d7cab9070362e8ee8f2acdff1188301fd569332eebffc91ac71bd41652
{
  "root": "0x",
  "status": "0x1",
  "cumulativeGasUsed": "0x2e260",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000080000000000000000000000000000000000",
  "logs": [
    {
      "address": "0xcc8650c9cd8d99b62375c22f270a803e7abf0de9",
      "topics": [
        "0xcde50e1bbc8495f4d015791042ac8d9b4e45d1cd60159e1fbad5863ee388d828"
      ],
      "data": "0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000077465737436363600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c646d732d656e64706f696e7400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000014616c7068612e7562696b6f6d2e63633a38383236000000000000000000000000",
      "blockNumber": "0x26f859",
      "transactionHash": "0x2e5089d7cab9070362e8ee8f2acdff1188301fd569332eebffc91ac71bd41652",
      "transactionIndex": "0x1",
      "blockHash": "0x00f4ed3b32a400eb5d66eb10f874c01318900fd02c4747ce9190a4f1beefb9af",
      "logIndex": "0x0",
      "removed": false
    }
  ],
  "transactionHash": "0x2e5089d7cab9070362e8ee8f2acdff1188301fd569332eebffc91ac71bd41652",
  "contractAddress": "0x0000000000000000000000000000000000000000",
  "gasUsed": "0xd115",
  "blockHash": "0x00f4ed3b32a400eb5d66eb10f874c01318900fd02c4747ce9190a4f1beefb9af",
  "blockNumber": "0x26f859",
  "transactionIndex": "0x1"
}
```

Here, we used gas-price flag to specify the gas price. Sometimes, the suggested gas price in Sepolia is too high (probably an issue with the 
new test network), and you might need to specify gas-price explicitly. If you don't, the suggested gas price value will be used.

To verify that the config value was updated, run:

```
ubikom-cli bc lookup config test666 --config-name=dms-endpoint --mode=test
18:55:41 WRN using Sepolia testnet
18:55:41 DBG using node node-url=https://sepolia.infura.io/v3/8f540714acb24862a8c9a5c3d8568f23
18:55:41 WRN using Sepolia testnet
18:55:41 DBG about to look up config config-name=dms-endpoint name=test666
alpha.ubikom.cc:8826
```
