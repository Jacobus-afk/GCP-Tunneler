#!/usr/bin/env bash

JQ_PREVIEW="jq --arg a ${2} -r '.[] | select( .project == \$a) | .backends.\"{}\".instance_groups[].instances[] ' ${1}"

jq --arg a ${2} -r ".[] | select( .project == \$a) | .backends[].name" ${1} | \
  fzf --tmux 80% --header-first --header $'Choose GCP backend\nCTRL-H for Previous Menu' \
  --preview "$JQ_PREVIEW" \
  --preview-label " Views " --bind 'ctrl-b:preview-up,ctrl-f:preview-down' \
  --bind "ctrl-h:print(**GO_BACK**)+accept" \
  --bind "enter:become($JQ_PREVIEW)"

