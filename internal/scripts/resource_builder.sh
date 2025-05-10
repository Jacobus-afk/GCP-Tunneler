#!/usr/bin/env bash

jq --arg a ${1} -c '.[] | select(.instances[$a]) | {project: .project} + .instances[$a]' instances.json

