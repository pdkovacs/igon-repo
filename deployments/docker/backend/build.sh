#!/bin/bash

docker_dir=deployments/docker/backend

set -x
  docker build -t iconrepo-backend:latest $docker_dir
set +x
