#!/bin/sh
# The first script which is ran to initialise the droplet.

set -e

unamestr=$(uname)
if [ "$unamestr" = 'Linux' ]; then
   sudo wget -O /usr/local/bin/copyback https://community-tools.sfo2.digitaloceanspaces.com/copyback_linux
   sudo wget -O /usr/local/bin/copyfrom https://community-tools.sfo2.digitaloceanspaces.com/copyfrom_linux
elif [ "$unamestr" = 'FreeBSD' ]; then
   sudo wget -O /usr/local/bin/copyback https://community-tools.sfo2.digitaloceanspaces.com/copyback_freebsd
   sudo wget -O /usr/local/bin/copyfrom https://community-tools.sfo2.digitaloceanspaces.com/copyfrom_freebsd
else
  echo "Unknown platform." 1>&2
  exit 1
fi

chmod 777 /usr/local/bin/copyfrom && chmod 777 /usr/local/bin/copyback
