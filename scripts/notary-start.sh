#!/bin/bash 

NOTARY=$GOPATH/src/github.com/theupdateframework/notary

(cd $NOTARY; docker-compose up -d)

docker run -d \
  --name registry \
  -p 5000:5000 \
  -v $NOTARY/fixtures:/certs \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/notary-server.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/notary-server.key \
  registry:2

docker ps
