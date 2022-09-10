export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n')

cd deployments/dev/keycloak && \
	docker-compose up --build -d && \
	./wait-for-local-keycloak.sh && \
	./create-terraform-client.sh && \
	cd terraform && \
	terraform init && \
	terraform apply -var "app_hostname=$MY_IP"
