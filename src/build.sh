#!/bin/sh

if [ "$1" = "-d" ]; then
    dpkg --add-architecture armhf
    apt update
    apt install -y gobjc-arm-linux-gnueabihf
    apt install -y libasound2-dev:armhf
fi

CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o ../oucher ./oucher.go
