#!/bin/bash -xe

# Mac OS: the following assumes you have executed
#     brew install coreutils
READLINK=$(greadlink --help >/dev/null 2>&1 && echo greadlink || echo readlink)

if [ "$0" = "bash" ];
then
    repo_root="$($READLINK -f $(dirname $BASH_SOURCE)/..)"
else
    repo_root="$($READLINK -f $(dirname $0)/..)"
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
    npm install \
        && npm run build:backend \
        && cp -a package*.json "$dist_dir/backend/" \
        && cp -a build/src/* "$dist_dir/backend/" \
    || return 1
    cd -
}

dist_frontend() {
    cd "${repo_root}/client"
    mkdir -p "$dist_dir/frontend/" 
    npm install \
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
#     "pwd && ls -al && node app.js",
# }
