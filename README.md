# Ubikom
*Encrypted email service based on decentralized private identity.*

## Identity

Your identity controls what you can do online. For most users it includes logins to multiple services, such as Google, Facebook, Amazon, etc. The commercial companies controlling such services have full control over your identity - they can restrict or disable it at any time, without explanation or recourse. 

The identity provided by such services also denies you privacy - their business model is to sell your private information to advertisers. Sure we can argue if their service is worth it, but arguably, for most of us, the answer to the question "how much of your privacy would you like to give up", the answer is, "none".

Our goal is to change this state of affairs, and give control over the user identity back to the user. It can then be used for authorization, secure communication, or to delegate some limited authority to act on behalf of the user to third party services. 

## Decentralized Identity Registry

Let's imagine instead a system where you control your identity, which means that you can prove to anyone that certain data originates from you. A common way to do it would be to have a private key and use it to sign outgoing messages. It can be also used to derive unique encryption keys so only the intended recipient of a message can read it.

Now let's assume that there is a registry where all the public keys are stored. Each key can be associated with a unique name. Whoever controls the corresponding private key, also controls the name.

This would give you something like a global user name - instead of having a single name and password per service, you can use your private key to control them all.

## Messaging

If you are like most people these days, you are constantly connected to the Internet, likely in more than one way. Considering this, it's somewhat unexpected that sending a simple message between two users is anything but simple. Sure, you can use GMail, or Twitter, or Discord - but why do you need help of some huge company just to send a message to your friend Bob? After all, the Internet is supposed to be a collection of public protocols enabling easy communication between individuals and organizations?

Let's see how private identity and the identity registry can help us here. With private keys controlled by users, and public keys available in the key registry, we encrypt messages so that only the intended recipient can read them.

With that in mind, message delivery becomes much easier - now we don't have to solve the problem of "how to deliver message to Bob and not somebody else". We are fine if the message ultimately makes it to Bob in whatever way, and we don't care who else can see it - since only Bob can decrypt it. For all we care, we can broadcast the message to the world - as long as Bob can receive it, we are good.

## Email

Which brings us to email, the biggest messaging system out there. It was supposed to be an easy way to send messages between users, and initially it was. However, the simplicity had its downside - because anyone could send messages to anyone else, without authentication or encryption, the abuse became common, and to confront this abuse, a patchwork of technologies was applied, making email increasingly more complex and unwieldy. 

Now it's virtually impossible to run your own email server, unless you want to spend significant time configuring and maintaining it.

Instead, we could use encryption and the identity registry to send emails between users in our system in a way that ensures privacy and prevents spam and abuse. Here, we have a working system that demonstrates this concept. 

## Architecture

For this prototype, we use the following architecture:

* *Identity service*, which is tasked with registering private keys, unique names and addresses
* *Dump service*, which accepts any valid messages address to any users, and stores them until the recipient comes to retrieve them. 
* *Email client proxy* which allows any email client to retrieve and send messages. The proxy handles the identity verification and encryption. 

### Aside - the "at" addresses

Every email address comes in bob@server format. Now that we talk about global identity, the "@server" part becomes obsolete - and it's no surprise that many services just use the username, or @bob handle. We have to comply with the email address format, but we want to de-emphasize the "at" part, so we will use the simplest possible format, bob@x. Yes, this is the user name followed by the "at" symbol and the letter x. We use x because it's cool, and because it looks like we just cancelled the whole service part - we put a cross in its place. 

## Testing the prototype

But enough talk, let's see how the prototype works. Before we do this, here's a necessary disclaimer:

**This is a prototype. It's not production-ready yet. Use it at your own risk. Things are likely to change in future, which includes changes breaking current functionality. The database can also be reset, wiping out whatever names you have registered.**

And another thing - in this example, we will use the public identity and dump server, but nothing prevents you from running your own (it's just another binary), so you will have your own private email system, where you and your buddies can exchange completely private messages. Not bad, huh? Just provide the new URL to the commands you run, and off you go. 

### Step 1: Get the binaries

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

### Step 2: Open the terminal and go to the directory where your binaries are located

For example, for Linux it would be (cloned repo location)/build/linux-amd64.

Go there, or add this directory to the path or whatever. 

### Step 3: Get your private key
To generate your private key, run the following command:

```
./ubikom-cli create key
```

This will create a new key and save it under $HOME/.ubikom/key

### Step 4: Register your key
Before you start using your private key, you need to register the public key (generated from the private key) with the service:

```
./ubikom-cli register key
```

This might take a few seconds, as the server requires you to generate proof-of-work and send it with the request, to prevent spam.

This is your private key, so yeah, don't give it to anyone.

### Step 5: Reserve your name
Name is what you use to send and receive messages. You must associate your name with your private key. The name must be unique - so if you get an error, try another name:

```
./ubikom-cli register name bob
```

Seriously, don't use bob with the public server. Bob is already registered. 

### Step 6: Start Ubikom Proxy
Proxy sits between your email client and the rest of Ubikom ecosystem, encrypting and signing your messages on the fly. 

```
./ubikom-proxy
```

***Known issue*** On Windows, you might get a "key not found" error, if you do, please specify the key location explicitly, like so:

```
ubikom-proxy --key=c:\users\bob\.ubikom\key
```

This issue will be fixed in the next release.

It will start and keep running, printing out a bunch of fun messages. For your mail client to work, the proxy must be running.

You can edit ubikom.conf file to configure the proxy, namely the ports and user names and passwords used by the SMTP and POP servers. 

Keep in mind that the data between email client and proxy is clear text. It's fine if you run it on your laptop, since the data never leaves your machine, but don't do it in multi-user environment. 

### Step 7: Configure your email client
I use [Mozilla Thunderbird](https://www.thunderbird.net/en-US/) for testing. It's great, and it's available for multiple platforms. Other might also work, so go ahead and try. 

Configure your email client as follows:

* Your email is "name@x", where name is the name you've registered in Step 6, and the rest is, yes, symbol @ and the letter x.
* Set your incoming mail server to POP3, localhost, port 1100
* Set your outgoing mail server to SMTP, localhost, port 1025
* For both POP and SMTP servers, the default name is "ubikom-user" and the password is "pumpkin123". Change it in ubikom.conf file, if you feel like it.

### Step 9: Use your mail client
Send secure email to other users, addressing it to name@x. 
Heck, send email to me, lgx@x! Definitely report bugs. 

## Current status

* Messages are encrypted and signed, which means only the intended recipient can read messages addressed to them.
* Sender and recipient names are not encrypted, so theoretically someone may find out that Bob sent message to Alice, but that's about it. You can't really do anything with names, since we must know who sent the message (to verify the signature), and who the recipient is.
* Eventually the identity service will be distributed and decentralized, but for now it's just a single machine.
* All messages are being sent via the big dump service, where they sit just as a bunch of bytes. You can run your own dump server (and identity server) instead.