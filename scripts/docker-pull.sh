#!/bin/bash

docker images -a | grep "hello-world" | awk '{print $3}' | xargs docker rmi -f

DOCKER_CONTENT_TRUST=1 DOCKER_CONTENT_TRUST_SERVER=https://localhost:4443 docker -D pull localhost:5000/hello-world:latest
