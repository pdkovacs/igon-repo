which ip && export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n') || export MY_IP=$(ifconfig | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}')

DOCKER_COMPOSE=docker-compose
which docker-compose >/dev/null 2>&1 || DOCKER_COMPOSE="docker compose"

cd deployments/dev/keycloak && \
	$DOCKER_COMPOSE up --build -d && \
	./wait-for-local-keycloak.sh && \
	./create-terraform-client.sh && \
	cd terraform && \
	terraform init && \
	terraform apply -var "app_hostname=$MY_IP"
