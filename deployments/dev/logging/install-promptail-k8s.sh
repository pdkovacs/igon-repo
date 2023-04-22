echo -n hostname: 
read hostname
echo -n My grafana.com API key:  
read -s my_grafana_api_key
curl -fsS https://raw.githubusercontent.com/grafana/loki/master/tools/promtail.sh | \
  sed "/      - args:/a\        - -client.external-labels=hostname=$hostname" | \
  sh -s 465842 $my_grafana_api_key logs-prod-012.grafana.net default | kubectl apply --namespace=default -f  -
