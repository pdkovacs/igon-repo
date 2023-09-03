dyndb_tf_dir=deployments/aws/dynamodb 

AWS_ACCESS_KEY_ID="$(terraform -chdir=$dyndb_tf_dir output -raw access_key_id)"
AWS_SECRET_ACCESS_KEY="$(terraform -chdir=$dyndb_tf_dir output -raw encrypted_access_key_secret | base64 --decode | keybase pgp decrypt)"

declare -A secrets
for secret_key in GITLAB_ACCESS_TOKEN OIDC_CLIENT_SECRET;
do
  secrets[$secret_key]="$(kubectl get secret iconrepo -ojsonpath='{.data.'$secret_key'}' | base64 --decode)"
  echo "$secret_key: ${secrets[$secret_key]}"
done

secrets[AWS_ACCESS_KEY_ID]=$AWS_ACCESS_KEY_ID
secrets[AWS_SECRET_ACCESS_KEY]=$AWS_SECRET_ACCESS_KEY

from_literal_args=""
for key in "${!secrets[@]}";
do
  value="${secrets[$key]}"
  echo "Key: $key"
  echo "Value: $value"
  from_literal_args="$from_literal_args --from-literal=$key=$value"
done

kubectl delete secret iconrepo 2>/dev/null || echo "Error while deleting secret 'iconrepo', probably didn't exist yet. Creating it..."
kubectl create secret generic iconrepo $from_literal_args
