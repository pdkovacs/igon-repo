#!/bin/bash

. ~/.keycloak.secrets
. ~/.iconrepo.secrets
ICONREPO_CLIENT_SECRET="$OIDC_CLIENT_SECRET"

bash ./wait-for-local-keycloak.sh
bash ./create-terraform-client.sh

# export TF_LOG=DEBUG
cd realm-client &&
  terraform init &&
  terraform apply -auto-approve \
    -var="tf_client_secret=$KEYCLOAK_TF_CLIENT_SECRET" \
    -var="client_secret=$ICONREPO_CLIENT_SECRET"

cd ../groups-users &&
  terraform init &&
  terraform apply -auto-approve \
    -var="tf_client_secret=$KEYCLOAK_TF_CLIENT_SECRET"
