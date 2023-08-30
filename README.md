# Ubikom Project
![GitHub Workflow Status](https://github.com/regnull/ubikom/actions/workflows/go.yml/badge.svg)
[![GoDoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/regnull/ubikom)
[![GoReportCard](https://goreportcard.com/badge/github.com/regnull/ubikom)](https://goreportcard.com/report/github.com/regnull/ubikom)
[![Coverage Status](https://coveralls.io/repos/github/regnull/ubikom/badge.svg?branch=master&foo=bar)](https://coveralls.io/github/regnull/ubikom?branch=master&foo=bar)

*Encrypted email service based on decentralized private identity.*

## Using Ubikom Email Service

Head over to [ubikom.cc](https://ubikom.cc) to create your identity and follow the instructions for email client setup.

## The Project

With Ubikom project, you can communicate via email in a secure way, while using the existing email clients that you know and love.

All email within Ubikom ecosystem is encrypted and authenticated.

There are no accounts. You create and register your private key, your possession of the private key is your identity.

You can run your own server, or you can interact with the public server. If you chose the latter, you temporary delegate the authority
to send and receive mail to the public proxy server. This delegation can be revoked at any time using your main private key.

You are also able to interact with the legacy email world using our gateway (coming up later).

## Why?

Long ago, you were able to run your own email server, which gave you an easy way to communicate with the world. Now you have to use Google or Microsoft for the simple task of sending messages to each other. Your identity is controlled by those companies, not by you. We want to give the identity back to the user and make it decentralized and not controlled by any entity. Based on this, we want to re-imagine email which is secure, private, and simple. It should be trivial for everyone (and everything) to register a name and start communicating.

## Getting the Binaries

As of now, you must run a few commands on your machine to generate the keys in a secure way. 

You can get binaries by compiling the source, or by pulling the pre-built binaries. The former is recommended, since you can examine the code to make sure no funny business is taking place. 

To compile the source, you must have Go and make installed.

To clone the repo, do:

```
git clone github.com/regnull/ubikom
```

Now build the binaries:

```
cd ubikom
make build
```

The binaries are placed in build directory, corresponding to your system (linux, windows or mac).

If you like to live dangerously, you can get the pre-build binaries by downloading the latest release from GitHub releases page.

## Using CLI

Ubikom CLI utility allows you to perform all low-level key, name and address operations. 

You can find the CLI command reference [here](https://github.com/regnull/ubikom/blob/master/doc/cli.md).

## References and Other Similar Projects

[Self-Sovereign Identity](https://en.wikipedia.org/wiki/Self-sovereign_identity)

[Decentralized Identifiers (DID)](https://www.w3.org/TR/did-core/)

[Sovrin - Global SSI](https://sovrin.org)

[In Search of Self-Sovereign Identity Leveraging Blockchain Technology](https://ieeexplore.ieee.org/document/8776589)

[The Path To Self-Sovereign Identity](http://www.lifewithalacrity.com/2016/04/the-path-to-self-soverereign-identity.html)

[EIDAS SUPPORTED SELF-SOVEREIGN IDENTITY](https://ec.europa.eu/futurium/en/system/files/ged/eidas_supported_ssi_may_2019_0.pdf)

[Blockchain and Digital Identity](https://www.eublockchainforum.eu/sites/default/files/report_identity_v0.9.4.pdf)

[SelfKey - SSI startup](https://selfkey.org)

[mnmnotmail](https://mnmnotmail.org)

[Re-thinking email](https://liw.fi/rethinking-email/)
