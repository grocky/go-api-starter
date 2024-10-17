#!/usr/bin/env bash

SOURCE_DIR=$(ag --go -l)

while sleep 1; do
  ag --go -l | entr -d -r $@
done

