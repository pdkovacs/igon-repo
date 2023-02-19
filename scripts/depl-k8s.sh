#!/bin/bash

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
  echo "kill_backend_process not implemented for k8s"
}

deploy_backend() {
  kubect apply -f ...
}

deploy_webpack_bundle() {
  kubect apply -f ...
}

# You can watch the app instances' outputs with something like this:
# for i in $(seq 0 $((app_instance_count -1))); do tilix -a session-add-down -x "tail -f $app_log$i" & ; done
