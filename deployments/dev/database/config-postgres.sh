#!/bin/bash

. ~/.postgres.secrets
kubectl delete secret postgres 2>/dev/null || echo "No postgres secret yet, creating..."
kubectl create secret generic postgres --from-literal=POSTGRES_PASSWORD=$POSTGRES_PASSWORD
