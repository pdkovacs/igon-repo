#!/bin/bash

cleanup() {
  rm keycloak-env.sh
}

cat > keycloak-env.sh <<EOF
KEYCLOAK_ADMIN=keycloak
KC_DB=postgres
KC_DB_URL_HOST=postgres
KC_DB_URL_PORT=5432
KC_DB_URL_DATABASE=keycloak
KC_DB_USERNAME=keycloak
EOF

trap cleanup EXIT

kubectl delete configmap keycloak || echo "No configmap for keycloak yet. Creating..."
kubectl create configmap keycloak --from-env-file=keycloak-env.sh
kubectl delete pod $(kubectl get pod -l app=keycloak -o jsonpath='{.items[0].metadata.name}') || echo "Keycloak is not running yet, nevermind..."

. ~/.keycloak.secrets
kubectl delete secret keycloak || echo "No secret for keycloak yet. Creating..."
kubectl create secret generic keycloak --from-literal=KEYCLOAK_ADMIN_PASSWORD=$KEYCLOAK_ADMIN_PASSWORD --from-literal=KC_DB_PASSWORD=$KC_DB_PASSWORD
