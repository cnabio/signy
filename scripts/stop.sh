#!/bin/bash

NOTARY=$GOPATH/src/github.com/theupdateframework/notary

(cd $NOTARY; docker-compose down)

docker stop registry
docker rm registry
rm -rf ~/.signy
rm -rf ~/.docker/trust/tuf/localhost:5000
docker ps