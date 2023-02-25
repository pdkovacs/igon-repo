#!/bin/bash

build_backend_docker() {
  executable=$1

	cp "$executable" deployments/docker/backend/
	eval "$(minikube docker-env)"
	deployments/docker/backend/build.sh
}

case $1 in
  build_backend_docker)
    build_backend_docker "$2"
  ;;
  *)
    echo "Unkown command $1";
    exit 1;
  ;;
esac
