# Ubikom
Free, secure and private messaging for everyone.

## Testing the early version
Early version means it's not entirely stable, so use your common sense.

### Step 1: Clone the Repo

```
git clone github.com/regnull/ubikom
```

### Step 2: Get your binaries
You can use pre-compiled binaries (under /build), or compile your own. To compile binaries, run /scripts/build.sh

### Step 3: Open the terminal and go to the directory where your binaries are located

For example, for Linix it would be (cloned repo location)/build/linux-amd64.

### Step 4: Get your private key
To generate your private key, run the following command:

```
./ubikom-cli create key
```

This will create a new key and save it under $HOME/.ubikom/key

### Step 5: Register your key
Before you start using your private key, you need to register the public key (generated from the private key) with the service:

```
./ubikom-cli register key
```

This might take a few seconds, as the server requires you to generate proof-of-work and send it with the request, to prevent spam.

### Step 6: Reserve your name
Name is what you use to send and receive messages. You must associate your name with your private key. The name must be unique - so if you get an error, try another name:

```
./ubikom-cli register name bob
```

### Step 7: Start Ubikom Proxy
Proxy sits between your email client and the rest of Ubikom ecosystem, encrypting and signing your messages on the fly. 

```
./ubikom-proxy
```
It will start and keep running, printing out a bunch of fun messages. For your mail client to work, the proxy must be running.

### Step 8: Configure your email client
I use [Mozilla Thunderbird](https://www.thunderbird.net/en-US/) for testing. Other might also work, so go ahead and try. 

Configure your email client as follows:

* Your email is "name@x", where name is the name you've registered in Step 6, and the rest is, yes, symbol @ and the letter x.
* Set your incoming mail server to POP3, localhost, port 1100
* Set your outgoing mail server to SMTP, localhost, port 1025
* User name and password can be anything, if the client asks you for password, you can enter any string.

### Step 9: Use your mail client
Send secure email to other users, addressing it to name@x. 
Heck, send email to me, lgx@x! Definitely report bugs. 