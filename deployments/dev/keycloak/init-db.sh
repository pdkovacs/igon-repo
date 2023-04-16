. ~/.keycloak.secrets

kubectl exec -i $(kubectl get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- psql -U postgres postgres <<EOF
create user keycloak with password '$KC_DB_PASSWORD';
alter user keycloak createdb;
EOF
kubectl exec -i $(kubectl get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- psql -U keycloak postgres <<EOF
create database keycloak;
EOF
