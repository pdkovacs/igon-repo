export ICON_REPO_CONFIG_FILE=deployments/dev/configurations/dev-oidc.json
cmd="make backend"
settle_down_secs=1

backend_executable="igo-repo"

project_dir="$(dirname $0)/.."
mkdir -p ~/workspace/logs
webpack_log=~/workspace/logs/igonrepo-webpack-build

pkill webpack
pkill "$backend_executable"

cd $project_dir/web
rm -rf $webpack_log
webpack --watch 2>&1 | tee $webpack_log &
cd -
while true;
do
    grep 'webpack.*compiled successfully' $webpack_log && break
    sleep 1
done

eval "$cmd" || exit 1
while true
do
  ./"$backend_executable" -l debug &
  sleep $settle_down_secs
  fswatch -r -1 --event Created --event Updated --event Removed -e '.*/[.]git/.*' -e 'web/node_modules' -e '.*/igo-repo/igo-repo$' .
  pkill "$backend_executable"
  eval "$cmd"
done
