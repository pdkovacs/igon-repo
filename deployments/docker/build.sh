#!/bin/bash

docker_dir=deployments/docker

set -x
cp -a igo-repo $docker_dir \
  && docker build -t iconrepo:latest $docker_dir \
  && docker tag iconrepo iconrepo:1.0.0
set +x
