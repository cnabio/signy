#!/bin/bash

brew install fswatch
make install

# https://emcrisostomo.github.io/fswatch/doc/1.14.0/fswatch.html/Tutorial-Introduction-to-fswatch.html#Detecting-File-System-Changes
# NOTE: We exclude bin/* to avoid infinite loop.
# TODO: Exclude *.sh, *.md, and other non-source files.
# FIXME: Sometimes fswatch fires a few times in a row. It is what it is.
fswatch -o . -e "bin/*" | (while read; do make install; date; echo; done)