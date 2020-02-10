#!/bin/bash

NOTARY=$GOPATH/src/github.com/theupdateframework/notary

(cd $NOTARY; docker-compose down)

docker stop registry
docker ps