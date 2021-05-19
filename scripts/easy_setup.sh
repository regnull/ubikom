#!/bin/bash

# Exit on any error.
set -e

CMD="./ubikom-cli"
DEFAULT_KEY_FILE=$HOME/.ubikom/key
DEFAULT_EMAIL_KEY_FILE=$HOME/.ubikom/email.key
DEFAULT_ADDRESS=alpha.ubikom.cc:8826
DEFAULT_PROXY=alpha.ubikom.cc

if [[ $# -ne 2 ]]; then
    echo 'Usage: easy_setup.sh name password'
    exit 1
fi

NAME=$1
PASSWORD=$2

echo "Checking if $NAME is available..."

# Check if the name is available.
LOOKUP_RESP=$($CMD lookup name $1 2>&1 | grep "not found")

if [ -z "$LOOKUP_RESP" ]; then 
    echo "$NAME is already registered, choose another name"
    exit 1
fi

if [ -e $DEFAULT_KEY_FILE ]; then
    echo "Private key $DEFAULT_KEY_FILE exists, delete or move to run this script"
    exit 1
fi

if [ -e $DEFAULT_EMAIL_KEY_FILE ]; then
    echo "Email key $DEFAULT_EMAIL_KEY_FILE exists, delete or move to run this script"
    exit 1
fi

echo 'Creating and registering the main key...'

$CMD create key
$CMD register key

echo 'Creating and registering the email key...'

SALT=$($CMD create key --out=$HOME/.ubikom/email.key --from-password=$PASSWORD 2>&1 | grep salt | sed 's/salt: //g')
if [ -z $SALT ]; then
    echo 'Error creating email key'
    exit 1
fi

$CMD register key --key=$HOME/.ubikom/email.key

echo 'Registering email key as child...'

$CMD register child-key --child=$HOME/.ubikom/email.key

echo 'Registering name and address...'

$CMD register name $NAME --target=$HOME/.ubikom/email.key
$CMD register address $NAME $DEFAULT_KEY_FILE --target=$HOME/.ubikom/email.key

echo 
echo "Use the following settings in your email client for POP/SMTP"
echo
echo "POP/SMTP server: $DEFAULT_PROXY"
echo "User name: $SALT"
echo "Password: $PASSWORD"