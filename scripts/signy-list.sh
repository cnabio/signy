#!/bin/bash

# FIXME: list does not seem to work right now
signy --tlscacert=$GOPATH/src/github.com/theupdateframework/notary/cmd/notary/root-ca.crt --server=https://localhost:4443 --log=info list localhost:5000/thin-bundle:v1