export ICON_REPO_CONFIG_FILE=deployments/configurations/dev.json
make build
count=0
while true
do
  count=$((count++))
  echo "${count}th start..."
  ./igo-repo -l debug &
  sleep 3
  fswatch -r -1 --event Created --event Updated --event Removed -e '.*/[.]git/.*' -e 'web/dist' -e 'web/node_modules' -e '.*/igo-repo/igo-repo$' .
  pkill igo-repo && wait
  make build
done
