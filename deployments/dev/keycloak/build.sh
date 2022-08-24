cd deployments/dev/keycloak && \
	docker-compose up --build -d && \
	./wait-for-local-keycloak.sh && \
	./create-terraform-client.sh && \
	cd terraform && \
	terraform init && \
	terraform apply
