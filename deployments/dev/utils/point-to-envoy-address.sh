kubectl get svc -n projectcontour | awk '
BEGIN {
  split("iconrepo iconrepo-backend", names);
}
/envoy/ {
  for ( name in names ) {
    printf("%s         %s.local.com\n", $3, names[name]);
  }
}
'
