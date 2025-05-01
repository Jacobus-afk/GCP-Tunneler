#!/usr/bin/env bash

jq --arg a ${2} -r ".[] | select( .project == \$a) | keys[] | select(. != \"project\" and . != \"zones\")" \
  ${1} | fzf --tmux 80% --header-first --header "Choose GCP compute view" \
  --preview "jq --arg a ${2} -r '.[] | select( .project == \$a) | .[\"{}\"][].name' ${1}" \
  --preview-label " Views " --bind 'ctrl-b:preview-up,ctrl-f:preview-down'

