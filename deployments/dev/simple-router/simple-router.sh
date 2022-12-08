my_parent_dir=$(dirname $0)
. $my_parent_dir/../../../scripts/functions.sh

export MY_IP=$(get_my_ip)
$my_parent_dir/../../../../simple-router/release/simple-router_darvin_amd64 \
  -c  hintAtUsername \
  -l $MY_IP:9999 \
  -d "http://$MY_IP:8091" \
  -r ".*alice@;http://$MY_IP:8091" \
  -r ".*joe@;http://$MY_IP:8092"
