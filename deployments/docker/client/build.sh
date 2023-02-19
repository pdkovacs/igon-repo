#!/bin/bash

docker_dir=deployments/docker/client

set -x
  docker build -t iconrepo-client:latest $docker_dir
set +x
