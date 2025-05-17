#!/usr/bin/env bash

jq --arg a ${2} -c '.[] | select(.instances[$a]) | {project: .project} + .instances[$a]' ${1}

