#!/bin/bash

depl_process_app_config() {
  if [ ! -f $ICON_REPO_CONFIG_FILE ];
  then
    echo "File $ICON_REPO_CONFIG_FILE doesn't exist"
    exit 1;
  fi

  kubectl create configmap iconrepo --from-file="$ICON_REPO_CONFIG_FILE" --dry-run=client -o yaml | kubectl apply -f -
}

kill_backend_process() {
  echo "kill_backend_process not implemented for k8s"
}

deploy_backend() {
  docker images -f reference='iconrepo:dev-*' --format=json | jq -r '.Tag' | \
  while read old_tag;
  do
    docker rmi "iconrepo:$old_tag"
  done
  scripts/make.sh build_backend_docker igo-repo
  tag=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13 ; echo '')
  new_dev_image=iconrepo:dev-"${tag,,}"
  docker tag iconrepo:latest "$new_dev_image" 
  kubectl set image deployment/iconrepo iconrepo="$new_dev_image"
}

deploy_webpack_bundle() {
  kubectl apply -f ...
}

# You can watch the app instances' outputs with something like this:
# for i in $(seq 0 $((app_instance_count -1))); do tilix -a session-add-down -x "tail -f $app_log$i" & ; done
