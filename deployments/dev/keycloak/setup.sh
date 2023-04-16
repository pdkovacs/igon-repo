#!/bin/bash

bash ./wait-for-local-keycloak.sh
bash ./create-terraform-client.sh

# export TF_LOG=DEBUG
cd realm-client && \
  terraform init && \
	terraform apply -auto-approve

cd ../groups-users && \
  terraform init && \
	terraform apply -auto-approve
