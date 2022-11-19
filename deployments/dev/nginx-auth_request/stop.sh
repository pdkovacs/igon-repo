my_parent_dir=$(dirname $0)
. $my_parent_dir/../../../scripts/functions.sh

export MY_IP=$(get_my_ip)

nginx_config_file=nginx.conf
export NGINX_CONFIG_FILE=$($READLINK -f $my_parent_dir/$nginx_config_file)

cd $my_parent_dir/docker
$DOCKER_COMPOSE -p nginx-auth_request down
