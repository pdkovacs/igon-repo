# Getting the test DNS names revolved
1. Hack your /etc/resolve.conf to look something like:
   ```
   # originally set to this one:
   # nameserver 172.25.144.1
   nameserver 127.0.0.1
   ```
1. Start dns mask with the config in this directory:
   ```
   sudo dnsmasq -k --conf-file=./dnsmasq.conf
   ```