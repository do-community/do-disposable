#!/usr/bin/env bash
# Heavily based off https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

mkdir dist

platforms=("windows/amd64" "windows/386" "windows/arm" "freebsd/amd64" "freebsd/386" "freebsd/arm" "freebsd/arm64" "linux/amd64" "linux/386" "linux/arm" "linux/arm64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='./dist/do-disposable_'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name .
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
