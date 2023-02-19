#!/bin/bash

export ICON_REPO_CONFIG_FILE=deployments/dev/app-configs/dev-oidc-simplerouter-template.json
export deployment_target=local # local or k8s

build_backend_cmd="make app"

build_backend

settle_down_secs=1

LOGS_HOME=~/workspace/logs
mkdir -p $LOGS_HOME
webpack_log=$LOGS_HOME/igonrepo-webpack-build

project_dir="$(dirname "$0")/.."
# shellcheck disable=SC1091
. "$project_dir/scripts/functions.sh"
. "$project_dir/scripts/depl-${deployment_target}.sh"

pkill webpack

depl_process_app_config
depl_kill_app_process

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
      echo "Client bundle recompiled, redeploying app..."
      deploy_webpack_bundle
    fi
  done
}

watch_backend() {
  eval "$build_backend_cmd" || exit 1
  while true
  do
    deploy_backend

    sleep $settle_down_secs
    fswatch -r -1 --event Created --event Updated --event Removed -e '.*/[.]git/.*' -e 'web' -e "$fswatch_pid_file"'$' -e '.*/igo-repo/igo-repo$' . &
    fswatch_pid=$!
    echo $fswatch_pid > "$fswatch_pid_file"
    wait $fswatch_pid
    [[ "$stopping" == "true" ]] && exit
    rm -rf "$fswatch_pid_file"
    
    kill_backend_process
    
    eval "$build_backend_cmd"
  done
}

# shellcheck disable=SC2164
cd "$project_dir/web"
echo "" > $webpack_log
watch_webpack &
npx webpack --watch 2>&1 | tee $webpack_log &
cd - || exit 1

watch_backend
