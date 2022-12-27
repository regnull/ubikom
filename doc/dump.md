# Dump Server

Dump server temporarily stores messages sent from one user to another. All 
messages are always encrypted, so we are less concerned with message security
in transit - we "dump" everything in one place, and then the recipient can
get the messages addressed to them, and only the recipient will be able
to decrypt the message.

Dump server interacts with the identity registry on the block chain. It 
performs two functions: send and receive, which work as follows:

* When a user sends a message, the dump server would use the
identity registry to verify the identity of the sender (make sure
that the private key they use corresponds to the public key in the
identity registry). Then, the server would accept the message and store
it in the local data directory. 
* When a user receives a message, they must first prove their identity
by presenting a valid signature. The server would return one message
per call, until there are no more messages.

## Running Dump Server

The easiest way to run dump server is as follows:

```
$ ubikom-dump --data-dir=some_directory --lookup-server-url=""
```

* You must specify --data-dir argument - this is where dump server stores
the encrypted messages.
* --lookup-server="" tells dump server to disable the legacy identity
registry lookups. This will go away later, when we finish transition to
Ethereum-based identity registry.

Some other flags to know:

--log-level controls logging (can be debug, info, warn, error);

--log-no-color disables fancy color log output (good when you save log to a file);

--network controls the Ethereum network dump server connects to. For now, the default 
is Sepolia test network,
to be changed to mainnet later. The valid arguments are "sepolia" (default), "main", 
or an explicit node address starting with "http://".

--contract-address defines the contract address on the blockchain - you probably don't need to change this one.

## Running Dump Server With Legacy Identity Registry

Going forward, the identity registry in Ethereum blockchain will be the only source
of truth. During the transition, however, we must support the clients registered
via the legacy identity registry. We actually have two legacy registries:

* The identity registry based on our own blockchain (this one is more recent).
* The identity registry implemented as a server (the older one).

To run the dump server with those as a fallback, specify --lookup-server-url like so:

```
$ ubikom-dump --data-dir=some_directory \
  --lookup-server-url="alpha.ubikom.cc:8825"
...
12/26 21:46:15 WRN using legacy lookup service url=alpha.ubikom.cc:8825
12/26 21:46:15 WRN using legacy blockchain url=http://18.223.40.196:8545
...
```

You will see the warnings in the logs. You will also see a warning each time 
dump server fails to resolve an identity using the Ethereum-based blockchain and
has to fallback on the legacy identity registry.