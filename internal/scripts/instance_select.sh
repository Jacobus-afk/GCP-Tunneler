#!/usr/bin/env bash

jq --arg a ${2} -r ".[] | select( .project == \$a) | .instances[].name" ${1} | \
  fzf -m --tmux 80% --header-first --header $'Choose GCP instance\nCTRL-H for Previous Menu, TAB/SHIFT-TAB to select/deselect multiple' \
  --bind "ctrl-h:print(**GO_BACK**)+accept"

