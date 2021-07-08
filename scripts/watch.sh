while true
do
  make build && ./igo-repo &
  sleep 10
  fswatch -xr -1 -m poll_monitor .
  pkill igo-repo && wait
done
