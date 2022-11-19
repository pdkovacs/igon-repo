my_parent_dir=$(dirname $0)
. $my_parent_dir/../../../scripts/functions.sh

export MY_IP=$(get_my_ip)

cd $my_parent_dir/

nginx_config_file=nginx.conf
cat  nginx.conf.template | sed -e 's/\$MY_IP/'$MY_IP'/g' > $nginx_config_file

export NGINX_CONFIG_FILE=$($READLINK -f $PWD/$nginx_config_file)
test -f $NGINX_CONFIG_FILE || (echo "NGINX_CONFIG_FILE $NGINX_CONFIG_FILE not found" && exit 1)

cd docker
$DOCKER_COMPOSE -p nginx-auth_request up
