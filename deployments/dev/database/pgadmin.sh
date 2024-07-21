. ~/.postgres.secrets

kubectl run --image dpage/pgadmin4 \
  --env PGADMIN_DEFAULT_EMAIL=peter.dunay.kovacs@gmail.com \
  --env PGADMIN_DEFAULT_PASSWORD="$PGADMIN_PASSWORD" \
  pgadmin

kubectl expose pod pgadmin --type=ClusterIP --port 80
