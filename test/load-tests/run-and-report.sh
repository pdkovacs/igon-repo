#!/bin/bash

__directory="$(readlink -f "$(dirname $0)")"

data_point_in_group_jql() {
  script_group="$1"
  metric="$2"
  echo ". | select(\
        .type==\"Point\" \
        and .metric == \"$metric\" \
        and .data.tags.group == \"$script_group\" \
    ) | .data.value"
}

count() {
  jq -s '{"count": length}'
}

avg() {
  jq -s '{"avg": (add / length)}'
}

min_p90_p95_max() {
  jq -s 'sort | .[0, (length * 0.9 | floor), (length * 0.95 | floor), length - 1]' |
    jq -s '{"min": .[0]},{"p90": .[1]},{"p95": .[2]},{"max": .[3]}'
}

stats() {
  script_group="$1"
  metric="$2"
  echo ""
  echo "===================================================================="
  echo "Statistics for $script_group"
  echo "===================================================================="
  echo "$metric"
  echo "--------------------------------------------------------------------"
  jq "$(data_point_in_group_jql "$script_group" "$metric")" output/create-icon-and-refresh.json | count
  jq "$(data_point_in_group_jql "$script_group" "$metric")" output/create-icon-and-refresh.json | avg
  jq "$(data_point_in_group_jql "$script_group" "$metric")" output/create-icon-and-refresh.json | min_p90_p95_max
}

mkdir -p output
k6 run \
  -e ICONREPO_BASE_URL=http://ux:ux@iconrepo.local.com \
  -e MY_SESION="$MY_SESION" \
  --vus 5 \
  --duration 30s \
  --out json=output/create-icon-and-refresh.json \
  "file://${__directory}/test-cases/create-icon-and-refresh.js"

stats "::Create and load image" "group_duration"
stats "::Create and load image::Create icon" "group_duration"
stats "::Create and load image::Load image" "group_duration"
