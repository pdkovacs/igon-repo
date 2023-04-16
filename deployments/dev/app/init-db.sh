read -s -p "Password: " password
kubectl exec -i $(kubectl get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- psql -U postgres postgres <<EOF
create user iconrepo with password 'iconrepo';
alter user iconrepo createdb;
EOF
kubectl exec -i $(kubectl get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- psql -U iconrepo postgres <<EOF
create database iconrepo;
EOF
