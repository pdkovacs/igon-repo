which ip && export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n') || export MY_IP=$(ifconfig | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}')

DOCKER_COMPOSE=docker-compose
which docker-compose >/dev/null 2>&1 || DOCKER_COMPOSE="docker compose"

cd deployments/dev/keycloak;

dcompose_project_in_list() {
	gawk '
BEGIN { projectNotFound = 1; }
/^NAME[[:blank:]]+STATUS[[:blank:]]+CONFIG FILES$/ { next; }
/^keycloak.*[/]deployments[/]dev[/]keycloak[/]docker-compose.yaml$/ { projectNotFound = 0; }
END { exit projectNotFound }
'
}

oldstate="$(shopt -po xtrace noglob errexit)"
set -e
which gawk > /dev/null 2>&1 || (echo "gawk is need for this" && exit 1);
set +vx; eval "$oldstate"

if $DOCKER_COMPOSE ls -a | dcompose_project_in_list;
then
	if $DOCKER_COMPOSE ls | dcompose_project_in_list;
	then
		exit 0
	fi
	$DOCKER_COMPOSE start
	exit 0
fi

$DOCKER_COMPOSE up --build -d && \
	./wait-for-local-keycloak.sh && \
	./create-terraform-client.sh && \
	cd terraform && \
	terraform init && \
	terraform apply -var "app_hostname=$MY_IP"
