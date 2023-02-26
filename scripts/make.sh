#!/bin/bash

build_backend_docker() {
  executable=$1

	cp "$executable" deployments/docker/backend/
	deployments/docker/backend/build.sh
}

build_client_docker() {
  ui_bundle="$1"
  ui_bundle_dir="$2"

  echo "[CLIENT] ui_bundle_dir: $ui_bundle_dir"

  cp "$ui_bundle" deployments/docker/client/
	cp "$ui_bundle_dir"/index.html deployments/docker/client/
	deployments/docker/client/build.sh
}

case $1 in
  build_backend_docker)
    build_backend_docker "$2"
  ;;
  build_client_docker)
    build_client_docker "$2"  "$3"
  ;;
  *)
    echo "Unkown command $1";
    exit 1;
  ;;
esac
