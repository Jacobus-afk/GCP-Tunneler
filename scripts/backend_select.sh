#!/usr/bin/env bash

jq --arg a ${2} -r ".[] | select( .project == \$a) | .backends[].name" ${1} | \
  fzf --tmux 80% --header-first --header $'Choose GCP backend\nSHIFT-TAB for Previous Menu' \
  --preview "jq --arg a ${2} -r '.[] | select( .project == \$a) | .backends.\"{}\".instance_groups[].instances[] ' ${1}" \
  --preview-label " Views " --bind 'ctrl-b:preview-up,ctrl-f:preview-down' \
  --bind "btab:print(**GO_BACK**)+accept"

