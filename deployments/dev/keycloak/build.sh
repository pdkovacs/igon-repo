which ip && export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n') || export MY_IP=$(ifconfig | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}')

cd deployments/dev/keycloak && \
	docker-compose up --build -d && \
	./wait-for-local-keycloak.sh && \
	./create-terraform-client.sh && \
	cd terraform && \
	terraform init && \
	terraform apply -var "app_hostname=$MY_IP"
