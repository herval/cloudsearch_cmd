#!/bin/sh

set -ex

go get github.com/GeertJohan/go.rice/rice
rice embed-go --import-path ./pkg/assets

go build -o cloudsearch ./cmd
