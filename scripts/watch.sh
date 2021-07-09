export ICON_REPO_CONFIG_FILE=deployments/configurations/dev.json
while true
do
  make build && ./igo-repo -l debug &
  sleep 3
  fswatch -xr -1 -m poll_monitor .
  pkill igo-repo && wait
done
