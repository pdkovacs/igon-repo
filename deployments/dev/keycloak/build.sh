#!/bin/bash

my_parent_dir=$(dirname $0)
. $my_parent_dir/../../functions.sh

export MY_IP=$(get_my_ip)

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
