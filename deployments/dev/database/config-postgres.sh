#!/bin/bash

[ -f ~/.postgres.secrets ] || echo "POSTGRES_PASSWORD=$(openssl rand -base64 32)" > ~/.postgres.secrets
. ~/.postgres.secrets
kubectl delete secret postgres 2>/dev/null || echo "No postgres secret yet, creating..."
kubectl create secret generic postgres --from-literal=POSTGRES_PASSWORD=$POSTGRES_PASSWORD --from-literal=PGADMIN_PASSWORD=$PGADMIN_PASSWORD
