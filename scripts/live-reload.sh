#!/bin/bash

TAG=latest

build () {
    echo "Building..."
    # Build an image containing python-in-toto to verify bundles/images with.
    docker build --rm -t cnabio/signy-in-toto-verifier:$TAG -f in-toto-container.Dockerfile .
    make TAG=$TAG install
    echo "...done."
    echo
}

echo "Installing fswatch..."
brew install fswatch
echo

build

# https://emcrisostomo.github.io/fswatch/doc/1.14.0/fswatch.html/Tutorial-Introduction-to-fswatch.html#Detecting-File-System-Changes
# NOTE: We exclude bin/* to avoid infinite loop.
# TODO: Exclude *.sh, *.md, and other non-source files.
# FIXME: Sometimes fswatch fires a few times in a row. It is what it is.
fswatch -o . -e "bin/*" | (while read; do build; done)