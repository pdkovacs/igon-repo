if [ -z "$APP_HOST" ];
then
    export MY_IP=$(ip route get 1.2.3.4 | awk '{print $7}' | tr -d '\n')
fi
export APP_HOST=$MY_IP # $MY_IP
export IDP_HOST=$MY_IP # localhost # 127.0.0.1

export SERVER_HOSTNAME="0.0.0.0"
export SERVER_PORT=8080
export SERVER_URL_CONTEXT="/icons"
export DB_HOST="$MY_IP"
export ICON_DATA_LOCATION_GIT="$PWD/tmp/data/git-repo"
export ICON_DATA_CREATE_NEW="init"
export AUTHENTICATION_TYPE="oidc"
export OIDC_CLIENT_ID="iconrepo-1"
export OIDC_CLIENT_SECRET="iconrepo-secret-1"
export OIDC_ACCESS_TOKEN_URL="http://$IDP_HOST:9001/token"
export OIDC_USER_AUTHORIZATION_URL="http://$IDP_HOST:9001/authorize"
export OIDC_CLIENT_REDIRECT_BACK_URL="http://$APP_HOST/icons/login"
export OIDC_TOKEN_ISSUER="http://$IDP_HOST:9001"
export OIDC_IP_JWT_PUBLIC_KEY_PEM_BASE64="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUFuRTR4bi9QMS9aemhpNm92QkFsZgpDUDN1MnNOeUswVjQ4RG1QTDFZU3FRSHZ6ZFhvMC80NEhEWjl2T1BCWFBBWjlPenRDeWJHaS81NjdRWVFac2pJClp2T3Ztcm9yNVkzL1hSZTZOQUVBL1hic3FsNTlDWjIrb1BDbVE5TlFHVk16bEEvK29VRnhJbUFWbnRZY2pCSysKZXdWVU4wM3hwcXkrcmk5dTFmbnNHVFZYRHRkalAxeDdJZWdUc2QxMEVocmJMcnhVbGcrZ29iTlZOUFIrZTV5dgo5azhQcXRzc1ZPUjBBamREeGtUazN3ODYwczRMVzAza3Blb05xODhaVEcyOE9MWWQ1eTNXRkowSjhlUDhtbkRXCmV3cEVudEpteGxZbWhOUVpQR091VjJoWm1pM21GeHZLeTFKNlNtSzB0MGNRNDlHbmNxZGNjK1JxS1VWSHJWSmsKVndJREFRQUIKLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
export OIDC_IP_LOGOUT_URL="http://$IDP_HOST:9001/logout"
export IGOREPO_LOG_LEVEL="debug"
export ICON_REPO_CONFIG_FILE=deployments/configurations/dev-oidc.json

cmd="make backend"
settle_down_secs=1

eval "$cmd" || exit 1
count=0
while true
do
  count=$((count++))
  echo "${count}th start..."
  ./iconrepo -l debug &
  sleep $settle_down_secs
  fswatch -r -1 \
    --event Created --event Updated --event Removed \
    -e '.*/[.]git/.*' \
    -e 'web/dist' \
    -e 'web/node_modules' \
    -e '.*/iconrepo/iconrepo$' \
    -e 'tmp/data' \
    .
  pkill iconrepo && wait
  eval "$cmd"
done
