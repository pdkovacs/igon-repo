#!/usr/bin/env bash

KEYCLOAK_URL="http://keycloak.internal/"

printf "Waiting for local Keycloak to be ready"

test_cmd="curl --output /dev/null --silent --head --fail --max-time 2 ${KEYCLOAK_URL}"

wait_in_loop() {
  local max_iteration_count=$1
  local iteration_count=0
  until test $iteration_count -ge $max_iteration_count;
  do
    if eval "$test_cmd";
    then
      return 0
    fi
    printf '.'
    iteration_count=$((iteration_count+1))
    sleep 2
  done
  return 1
}

if ! wait_in_loop 5;
then
  printf '\n'
  echo "Is minikube tunnel started?"
  if ! wait_in_loop 5
  then
    echo "Is keycloak external IP mapped to the \"keycloak\" DNS name?"
  fi
  wait_in_loop 100000000
fi
