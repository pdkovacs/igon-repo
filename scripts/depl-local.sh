#!/bin/bash

app_instance_count=2
app_executable="igo-repo"

depl_process_app_config() {
  if echo "$ICON_REPO_CONFIG_FILE" | grep -E '\-template.json';
  then
    # shellcheck disable=SC2001
    NEW_ICON_REPO_CONFIG_FILE=$(echo "$ICON_REPO_CONFIG_FILE" | sed -e 's/^\(.*\)-template[.]json$/\1.json/g')
    echo "$NEW_ICON_REPO_CONFIG_FILE"
    envsubst < "$ICON_REPO_CONFIG_FILE" > "$NEW_ICON_REPO_CONFIG_FILE"
    export ICON_REPO_CONFIG_FILE=$NEW_ICON_REPO_CONFIG_FILE
    echo "ICON_REPO_CONFIG_FILE is $ICON_REPO_CONFIG_FILE"
  fi
}

kill_backend_process() {
  pkill "$app_executable"
  pkill "$app_executable"
}

deploy_app() {
  logs_home=${LOGS_HOME:-"${HOME}/workspace/logs"}
  app_log=$logs_home/iconrepo-app-
  logfiles=""
  
  for i in $(seq 0 $((app_instance_count -1)));
  do
    export SERVER_PORT=$((8091 + "$i"))
    LOAD_BALANCER_ADDRESS=$(get_my_ip):9999
    export LOAD_BALANCER_ADDRESS
    ./"$app_executable" -l debug >"$app_log$i" 2>&1 &
    logfiles="$logfiles
    tail -f $app_log$i"
  done
  echo "$logfiles"
}

get_fswatch_pid() {
  while ! test -f "$fswatch_pid_file"; do
    sleep 1
    echo "Still waiting for $fswatch_pid_file..."
  done
  cat "$fswatch_pid_file"
}

deploy_webpack_bundle() {
  fswatch_pid=$(get_fswatch_pid)
  echo "Killing fswatch with pid $fswatch_pid..."
  kill "$fswatch_pid"
}

# You can watch the app instances' outputs with something like this:
# for i in $(seq 0 $((app_instance_count -1))); do tilix -a session-add-down -x "tail -f $app_log$i" & ; done
