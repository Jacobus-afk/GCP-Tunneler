#!/usr/bin/env bash

jq -r '.[] | .project' ${1} | fzf --tmux 80% --header-first --header "Choose GCP project" \
--preview "jq --arg a {} -r '.[] | select( .project == \$a) | .instances[].name' ${1}" \
--preview-label " Instances " --bind 'ctrl-b:preview-up,ctrl-f:preview-down'
