#!/bin/bash 

NOTARY=$GOPATH/src/github.com/theupdateframework/notary

(cd $NOTARY; docker-compose up -d)

# NOTE: Notary (see scripts/notary-start.sh) seems to require TLS for both the
# Registry and itself. However, that setup breaks cnab-to-oci (required for
# signy), most likely because we use a self-signed root here. Until we fix
# this, it is easiest to use two different scripts to initalize the Registry
# for Notary and signy.
docker run -d \
  --name registry \
  -p 5000:5000 \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
  registry:2

docker ps
