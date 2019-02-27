#!/bin/sh

go build -o cloudsearch ./cmd
GOOS=linux GOARCH=arm go build -o cloudsearch.bin ./cmd
GOOS=windows GOARCH=386 go build -o cloudsearch.exe ./cmd