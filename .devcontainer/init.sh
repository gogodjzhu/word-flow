#!/bin/bash

# check if the script is run as root
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

# Customize configuration
echo "export HISTSIZE=10000000" >> /etc/profile

# check if apt-get is installed
if ! [ -x "$(command -v apt-get)" ]; then
  echo 'Error: apt-get is not installed.' >&2
  exit 1
fi

# Prepare package
apt-get update -y
apt-get install -y curl wget iputils-ping net-tools vim
curl -fsSL https://opencode.ai/install | bash

# git config
git config --global --add safe.directory '*'

# set go env
go mod tidy
go mod vendor