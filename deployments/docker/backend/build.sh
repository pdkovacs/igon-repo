#!/bin/bash

docker_dir=deployments/docker/backend

set -x
  docker build -t iconrepo:latest $docker_dir
set +x
