#!/bin/bash

# Clone Notary.
(cd /tmp; go get github.com/theupdateframework/notary)

# We will sign and push this to our localhost Notary and Registry.
docker pull hello-world
docker tag hello-world localhost:5000/hello-world
docker images