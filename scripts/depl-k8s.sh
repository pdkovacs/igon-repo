#!/bin/bash

deploy_app_config() {
  if [ ! -f "$ICON_REPO_CONFIG_FILE" ];
  then
    echo "File $ICON_REPO_CONFIG_FILE doesn't exist"
    exit 1;
  fi

  cp "$ICON_REPO_CONFIG_FILE" deployments/dev/app/config.json

  kubectl create configmap iconrepo --from-file=deployments/dev/app/config.json --dry-run=client -o yaml | kubectl apply -f -
}

deploy_client_config() {
  kubectl create configmap iconrepo-client --from-file=deployments/dev/app/client-config.json --dry-run=client -o yaml | kubectl apply -f -
}

kill_backend_process() {
  echo "kill_backend_process not implemented for k8s"
}

build_backend() {
  GOOS=linux GOARCH=amd64 make app
}

build_image() {
  build_image_cmd="$1"
  (
    eval $(minikube docker-env)
    eval "$build_image_cmd"
  )
}

redeploy_service() {
  image_name="$1"
  build_image_cmd="$2"
  deployment_name="$3"
  container_name="$4"

  docker images -f reference="$image_name"':dev-*' --format=json | jq -r '.Tag' | \
  while read -r old_tag;
  do
    docker rmi "$image_name:$old_tag"
  done
  build_image "$build_image_cmd"
  tag=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13 ; echo '')
  new_dev_image="$image_name":dev-"${tag,,}"
  docker tag "$image_name":latest "$new_dev_image" 
  kubectl set image deployment/"$deployment_name" "$container_name"="$new_dev_image"
}

deploy_backend() {
  deploy_app_config
  build_image_cmd="scripts/make.sh build_backend_docker igo-repo"
  # redeploy_service iconrepo-backend "$build_image_cmd" iconrepo iconrepo
  build_image "$build_image_cmd"
  kubectl delete pod $(kubectl get pod -l app=iconrepo -o jsonpath='{.items[0].metadata.name}')
}

deploy_webpack_bundle() {
  ui_bundle="$1"
  ui_bundle_dir="$2"
  deploy_client_config
  build_image_cmd="scripts/make.sh build_client_docker \"$ui_bundle\" \"$ui_bundle_dir\""
  set -x
  # redeploy_service iconrepo-client "$build_image_cmd" iconrepo-client iconrepo-client
  build_image "$build_image_cmd"
  kubectl delete pod $(kubectl get pod -l app=iconrepo -o jsonpath='{.items[0].metadata.name}')
  set +x
}

# You can watch the app instances' outputs with something like this:
# for i in $(seq 0 $((app_instance_count -1))); do tilix -a session-add-down -x "tail -f $app_log$i" & ; done
