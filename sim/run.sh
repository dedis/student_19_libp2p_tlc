#!/bin/sh

cd protocol
GOOS=linux go build
cd ..
go build
./sim -platform deterlab run.toml