# Configuration for contour.io's HTTPProxy custom resource

1. Install the custom resource definition:
   
   ```
   kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
   ```

2. Install the manifests for the app:
   
   ```
   kubectl apply -f cors/backend-httpproxy.yaml
   kubectl apply -f cors/client-httpproxy.yaml
   ```

3. Check and note the FQDNs:
   
    ```
    $ kubectl get httpproxies.projectcontour.io 
    NAME                       FQDN                         TLS SECRET   STATUS   STATUS DESCRIPTION
    content-delivery-ingress   iconrepo.local.com                        valid    Valid HTTPProxy
    iconrepo-backend-ingress   iconrepo-backend.local.com                valid    Valid HTTPProxy
    ```

4. Check and note envoy's IP address
   
    ```
    $ minikube service -n projectcontour envoy --url
    http://192.168.64.4:31419
    http://192.168.64.4:32056
    ```

5. Add the mapping of the FQDNs to envoy's IP address in your host environment (e.g. in `/etc/hosts`)
