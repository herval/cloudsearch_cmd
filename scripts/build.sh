#!/bin/sh

set -ex

go get github.com/GeertJohan/go.rice/rice
rice embed-go

go build -o cloudsearch ./cmd
