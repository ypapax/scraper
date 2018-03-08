#!/usr/bin/env bash
set -ex
cd $GOPATH/src/github.com/ypapax/scraper
go install && scraper -url http://$1 -from $2 -to $3