#!/usr/bin/env bash

jq -r '.[] | .project' instances.json | fzf --tmux 80% --header-first --header "Choose GCP instance" \
--preview "jq --arg a {} -r '.[] | select( .project == \$a) | .instances[].name' instances.json" \
--bind 'ctrl-b:preview-up,ctrl-f:preview-down'
