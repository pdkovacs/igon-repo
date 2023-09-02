#!/bin/bash -xe

# Mac OS: the following assumes you have executed
#     brew install coreutils
READLINK=$(which greadlink >/dev/null 2>&1 && echo greadlink || echo readlink)
DATE=$(which gdate >/dev/null 2>&1 && echo gdate || echo date)

get_my_ip() {
  if which ip >/dev/null;
  then
    ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n'
  else
    ifconfig | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}'
  fi
}

MY_IP="$(get_my_ip)"
export MY_IP

export DOCKER_COMPOSE=docker-compose
which docker-compose >/dev/null 2>&1 || export DOCKER_COMPOSE="docker compose"

if [ "$0" = "bash" ];
then
    repo_root="$($READLINK -f "$(dirname "${BASH_SOURCE[0]}")"/..)"
else
    repo_root="$($READLINK -f "$(dirname "$0")/..")"
fi

dist_dir="${repo_root}/deploy/dist"

clean_backend_dist() {
    rm -rf backend/build \
    && rm -rf backend/node_modules \
    && rm -rf backend/package-lock.json
}

clean_client_dist() {
    rm -rf client/dist \
    && rm -rf client/node_modules \
    && rm -rf client/package-lock.json
}

clean_dist() {
    rm -rf "$dist_dir" && clean_backend_dist && clean_client_dist
}

dist_backend() {
    cd "${repo_root}/backend"
    mkdir -p "$dist_dir/backend/"
    npm ci \
        && npm run build:backend \
        && cp -a package*.json "$dist_dir/backend/" \
        && cp -a build/src/* "$dist_dir/backend/" \
    || return 1
    cd -
}

dist_frontend() {
    cd "${repo_root}/client"
    mkdir -p "$dist_dir/frontend/" 
    npm ci \
        && rm -rf dist \
        && npm run dist \
        && cp -a "${repo_root}"/client/dist/* "$dist_dir/frontend/" \
    || return 1
    cd -
}

dist_all() {
    clean_dist && dist_backend && dist_frontend
}

pack() {
    dist_all \
    && cd "$dist_dir" \
    && tar -czf icon-repo-app.tgz --exclude '.DS_Store' backend frontend
}

# build_docker() {
#     dist_all && cd docker && docker build -t cxn/icon-repo-app .
# }

# start_docker() {
#     pwd && ls -al && node app.js
# }

ns_date() {
  "$DATE" --rfc-3339 ns
}
