#!/bin/bash

export ICON_REPO_CONFIG_FILE=deployments/dev/app/dev-oidc.json
export deployment_target=k8s # "local" or "k8s"
export backend_client_split=cdn-origin # "cors" or "cdn-origin"


ui_bundle="$1"
ui_bundle_dir="$2"

settle_down_secs=1

LOGS_HOME=~/workspace/logs
mkdir -p $LOGS_HOME
webpack_log=$LOGS_HOME/igonrepo-webpack-build

project_dir="$(dirname "$0")/.."
#shellcheck disable=SC2164
. "$project_dir/scripts/functions.sh"
. "$project_dir/scripts/depl-${deployment_target}.sh"

pkill webpack

kill_backend_process

tail_command_pattern="[t]ail.*${webpack_log}"
ps -ef | awk -v tail_command_pattern="$tail_command_pattern" '$0 ~ tail_command_pattern { print $0; system("kill " $2); }'

unset stopping

cleanup() {
  echo "Cleaning up $$..."
  stopping=true
  ps -ef | awk -v tail_command_pattern="$tail_command_pattern" '$0 ~ tail_command_pattern { print $0; system("kill " $2); }'
  pkill -9 -P $$
}

trap cleanup EXIT SIGINT SIGTERM

fswatch_pid_file="$($READLINK -f "$project_dir/fswatch.pid")"

watch_webpack() {
  tail -F -n5000 $webpack_log | while IFS= read -r line;
  do
    if echo "$line" | grep 'webpack.*compiled successfully';
    then
      echo "[CLIENT] Client bundle recompiled, redeploying app..."
      deploy_webpack_bundle "$ui_bundle" "$ui_bundle_dir"
      ns_date
    fi
  done
}

watch_backend() {
  build_backend || exit 1
  while true
  do
    deploy_backend

    sleep $settle_down_secs

set -x
    fswatch -r -1 --event Created --event Updated --event Removed \
      -e '.*/[.]git/.*' \
      -e 'web' \
      -e "$fswatch_pid_file"'$' \
      -e '.*/igo-repo/igo-repo$' \
      -e 'deployments' \
      . &
set +x

    fswatch_pid=$!
    echo $fswatch_pid > "$fswatch_pid_file"

    ns_date
    wait $fswatch_pid

    [[ "$stopping" == "true" ]] && exit
    rm -rf "$fswatch_pid_file"
    
    kill_backend_process
    
    build_backend
  done
}

echo "" > $webpack_log
watch_webpack &
#shellcheck disable=SC2164
cd "$project_dir/web"
npx webpack --watch 2>&1 | tee $webpack_log &
cd - || exit 1

set -x
watch_backend
set +x
