#!/bin/sh

mkdir /tmp/camlistore
export CAMLI_PASSWORD=foo
make && ./camlistore