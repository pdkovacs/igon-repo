#!/bin/bash

. ~/.iconrepo.secrets
kubectl delete secret iconrepo 2>/dev/null || echo "Error while deleting secret 'iconrepo', probably didn't exist yet. Creating it..."
kubectl create secret generic iconrepo \
  --from-literal=OIDC_CLIENT_SECRET=$OIDC_CLIENT_SECRET \
  --from-literal=GITLAB_ACCESS_TOKEN=$GITLAB_ACCESS_TOKEN \
  --from-literal=AWS_ACCESS_KEY_ID="" \
  --from-literal=AWS_SECRET_ACCESS_KEY=""
kubectl delete pod $(kubectl get pod -l app=iconrepo -o jsonpath='{.items[0].metadata.name}')
