if [ -z "$MY_IP" ];
then
    export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n')
fi

export APP_HOST=$MY_IP
export IDP_HOST=$MY_IP # 127.0.0.1 # $MY_IP

export TMP_CONFIG_ROOT=$PWD/tmp/config
export OIDC_CONFIG_FILE=$TMP_CONFIG_ROOT/oidc-config.json
export NGINX_CONFIG_FILE=$TMP_CONFIG_ROOT/nginx.conf
export ICONREPO_DATA=$TMP_CONFIG_ROOT/iconrepo-data
export ICONREPO_CONFIG_FILE=$ICONREPO_DATA/config.json

mkdir -p $TMP_CONFIG_ROOT
mkdir -p $ICONREPO_DATA

create_oidc_config() {
    cat > $OIDC_CONFIG_FILE <<EOF
{
    "clients": [
        {
            "client_id": "iconrepo-1",
            "client_secret": "iconrepo-secret-1",
            "redirect_uris": [
                "http://$IDP_HOST/icons/login"
            ],
            "scope": "openid profile email phone address"
        }
    ]
}
EOF
}

create_nginx_config() {
    envsubst < nginx.conf.template > $NGINX_CONFIG_FILE
}

create_iconrepo_config() {
    cat > $ICONREPO_CONFIG_FILE <<EOF
{
    "usersByRoles": {
        "ICON_EDITOR": [
            "alice.wonderland@example.com"
        ]
    }
}
EOF
}

create_oidc_config
create_nginx_config
create_iconrepo_config

docker-compose up
