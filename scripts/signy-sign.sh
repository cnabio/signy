#!/bin/bash

source scripts/signy-env.sh

signy --tlscacert=$GOPATH/src/github.com/theupdateframework/notary/cmd/notary/root-ca.crt --server=https://localhost:4443 --log=info sign testdata/cnab/bundle.json localhost:5000/helloworld-thin-bundle:v1