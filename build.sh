#!/usr/bin/env bash
# Heavily based off https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

# 1) Build the do-disposable binary

mkdir cli-dist

cp do-disposable-install.ps1 cli-dist/do-disposable-install.ps1
cp do-disposable-install.sh cli-dist/do-disposable-install.sh

platforms=("windows/amd64" "windows/386" "windows/arm" "freebsd/amd64" "freebsd/386" "freebsd/arm" "linux/amd64" "linux/386" "linux/arm" "linux/arm64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='./cli-dist/do-disposable_'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name .
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

# 2) Build the copyback/copyfrom binaries

platforms=("freebsd/amd64" "linux/amd64")

mkdir droplet-tools-dist

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='../droplet-tools-dist/copyback_'$GOOS

    cd copyback
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name .
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    cd ..

    output_name='../droplet-tools-dist/copyfrom_'$GOOS
    cd copyback
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name .
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    cd ..
done
