#!/bin/bash

users_initializer_image_name=init-keycloak-users

eval "$(minikube docker-env)"
docker build -t $users_initializer_image_name .

kubectl run $users_initializer_image_name \
  --image=$users_initializer_image_name \
  --rm -i -t \
  --restart=Never \
  --image-pull-policy=Never
