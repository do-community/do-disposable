#!/bin/sh
set -e

arch=$(uname -m)
if [ "$arch" = 'x86_64' ]; then
    arch="amd64"
elif [ "$arch" = 'i386' ] || [ "$arch" = "i686" ]; then
    arch="386"
elif [ "$arch" = armv7* ]; then
    arch="arm"
elif [ "$arch" = armv8* ]; then
    arch="arm64"
else
  echo "Unknown architecture." 1>&2
  exit 1
fi

unamestr=$(uname)
if [ "$unamestr" = 'Linux' ]; then
    sudo wget -O /usr/local/bin/do-disposable https://github.com/do-community/do-disposable/releases/download/v1.0.0/do-disposable_linux-$arch
elif [ "$unamestr" = 'FreeBSD' ]; then
    sudo wget -O /usr/local/bin/do-disposable https://github.com/do-community/do-disposable/releases/download/v1.0.0/do-disposable_freebsd-$arch
elif [ "$unamestr" = 'Darwin' ]; then
    sudo wget -O /usr/local/bin/do-disposable https://github.com/do-community/do-disposable/releases/download/v1.0.0/do-disposable_darwin-$arch
else
  echo "Unknown platform." 1>&2
  exit 1
fi

sudo chmod +x /usr/local/bin/do-disposable
