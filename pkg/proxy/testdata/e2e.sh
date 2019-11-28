#!/bin/sh
set -eux

export GOOS=linux
export CGO_ENABLED=0

cd ../..
go build -ldflags='-s -w'
mv lambdahttp bootstrap

cd pkg/proxy/testdata
go build -ldflags='-s -w'
mv testdata ../../../hello.handler
cd -

zip lambda.zip bootstrap hello.handler
