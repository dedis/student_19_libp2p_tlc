#!/bin/sh

cd protocol
GOOS=linux go build
cd ..
go build
./sim -platform deterlab -runwait 60m0s -experimentwait 60m0s  run.toml | tee log.log
# grep -E "[0-9]{10}" log.log > logDeterlab.log
