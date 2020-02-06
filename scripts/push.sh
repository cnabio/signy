#!/bin/bash
DOCKER_CONTENT_TRUST=1 DOCKER_CONTENT_TRUST_SERVER=https://localhost:4443 docker -D push localhost:5000/hello-world:latest
