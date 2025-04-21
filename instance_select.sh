#!/usr/bin/env bash

jq --arg a ${2} -r ".[] | select( .project == \$a) | .instances[].name" ${1} | \
  fzf --tmux 80% --header-first --header "Choose GCP instance"

