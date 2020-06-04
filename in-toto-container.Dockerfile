# Choose a base image with a larges number of packages out of the box, so that
# in-toto inspections containing arbitrary commands are likely to succeed.
FROM ubuntu:latest

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y python3-pip \
    && apt-get autoremove \
    && apt-get autoclean \
    && pip3 --no-cache install in-toto \
    # A directory where we will copy all links, layouts, and pubkeys.
    && mkdir /in-toto \
    # Let bash figure out what the root layout and its pubkeys are called.
    && echo 'in-toto-verify --layout *.layout --layout-keys *.pub --link-dir . --verbose' > /in-toto/verify.sh

ENTRYPOINT ["bash", "/in-toto/verify.sh"]