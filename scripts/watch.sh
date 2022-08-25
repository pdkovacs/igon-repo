#!/bin/bash

export ICON_REPO_CONFIG_FILE=deployments/dev/configurations/dev-oidc.json
cmd="make backend"
settle_down_secs=1

backend_executable="igo-repo"

READLINK=greadlink
which $READLINK || READLINK=readlink

project_dir="$($READLINK -f $(dirname $0)/..)"
echo "Project dir: $project_dir"
mkdir -p ~/workspace/logs
webpack_log=~/workspace/logs/igonrepo-webpack-build

pkill webpack
pkill "$backend_executable"

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

fswatch_pid_file="$project_dir/fswatch.pid"

get_fswatch_pid() {
  while ! test -f "$fswatch_pid_file"; do
    sleep 1
    echo "Still waiting for $fswatch_pid_file..."
  done
  cat $fswatch_pid_file  
}

watch_webpack() {
  tail -F -n5000 $webpack_log | while IFS= read -r line;
  do
    if echo $line | grep 'webpack.*compiled successfully';
    then
      fswatch_pid=$(get_fswatch_pid)
      echo "Client bundle recompiled, restarting backend (pid: $fswatch_pid)..."
      kill $fswatch_pid
    fi
  done
}

watch_backend() {
  eval "$cmd" || exit 1
  while true
  do
    ./"$backend_executable" -l debug &
    sleep $settle_down_secs
    set -x
    fswatch -r -1 --event Created --event Updated --event Removed -e '.*/[.]git/.*' -e 'web' -e $fswatch_pid_file'$' -e '.*/igo-repo/igo-repo$' . &
    fswatch_pid=$!
    echo $fswatch_pid > "$fswatch_pid_file"
    wait $fswatch_pid
    [[ "$stopping" == "true" ]] && exit
    rm -rf $fswatch_pid_file
    set +x
    pkill "$backend_executable"
    eval "$cmd"
  done
}

cd $project_dir/web
echo "" > $webpack_log
watch_webpack &
webpack --watch 2>&1 | tee $webpack_log &
cd -

watch_backend
