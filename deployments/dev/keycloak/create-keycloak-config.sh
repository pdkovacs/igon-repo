#!/bin/bash

cleanup() {
  rm keycloak-env.sh
}

cat > keycloak-env.sh <<EOF
KEYCLOAK_ADMIN=keycloak
KEYCLOAK_ADMIN_PASSWORD=password
KC_DB=postgres
KC_DB_URL_HOST=postgres
KC_DB_URL_PORT=5432
KC_DB_URL_DATABASE=keycloak
KC_DB_USERNAME=keycloak
KC_DB_PASSWORD=keycloak
EOF

trap cleanup EXIT

kubectl delete configmap keycloak
kubectl create configmap keycloak --from-env-file=keycloak-env.sh
kubectl delete pod $(kubectl get pod -l app=keycloak -o jsonpath='{.items[0].metadata.name}')
