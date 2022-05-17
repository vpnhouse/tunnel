# Configuration file reference

Default path is `/opt/vpnhouse/tunnel/config.yaml`.

```yaml
# config.yaml
log_level: debug
sqlite_path: /opt/vpnhouse/tunnel/db.sqlite3

# serve openAPI documentation under the `/rapidoc/` path if enabled
# https://mrin9.github.io/RapiDoc/
rapidoc: true

# configuration of the HTTP default server
http:
  listen_addr: "0.0.0.0:80"
  # enable CORS for local development
  # optional, default: false
  cors: false
  # expose prometheus counters on /metrics
  prometheus: true
 
# we can also serve SSL traffic with valid certificates by LetsEncrypt.
# Please take a look at the section `domain` below.
ssl:
  listen_addr: :443
  
# how this instance may be referenced by the DNS name?
domain:
  # mode can be "direct" or "reverse-proxy".
  # direct means that the instance handle the internet traffic by its own,
  # thus it also must serve the SSL certificate (if any).
  # the "reverse-proxy" mode means that there is a web server (nginx, apache, traefic, etc) 
  # in front of the instance, so this web server handles SSL traffic, so we don't have to
  # do anything with it.
  mode: "direct" # or "reverse-proxy"
  # domain name, in case of mode=direct we'll try to issue the SSL certificate for this name.
  name: "xxx.themachine.org"
  # should we issue the certificate? works only with "mode=direct",
  # requires setting the `ssl.listen_addr` option (see above).
  issue_ssl: true
  # where to store issued certificate bundles,
  # equals to the configuration directory in most cases.
  dir: /opt/vpnhouse/tunnel

wireguard:
    # interface name, the interface will be allocated automatically
    interface: "uwg0"
    # public IPv4 of the server, announced to clients
    server_ipv4: "1.2.3.4"
    # Public UDP port of a wireguard server.
    # This value is announced to peers, in 99% cases it is the same as the `nated_port`.
    # May differs from the `nated_port`'s value if NATed (especially with docker).
    server_port: 3000
    # Wireguard listen port inside the container.
    # In 99% cases it matches the `server_port` value.
    nated_port: 3000
    # keepalive interval
    keepalive: 60
    # subnet for VPN clients, server will take the first available address automatically
    subnet: "10.235.0.0/24"
    # a list of DNS servers to announce to clients
    dns:
        - 8.8.8.8
        - 8.8.4.4
    # wireguard private key, generated automatically on the first start 
    private_key: 4BsYp8MzCvIgIwQrHIj9LW7Njrq4QoM1BR7HNC/1j1k=
    
network:
  access:
    # can be "internet_only" or "allow_all",
    # the latter allows peers to talk to each other
    # like in a normal old-school LANs. 
    default_policy: "internet_only"
  rate_limit:
    # total available bandwidth, used for packets scheduling via tc.
    # use k or K for Kbps,
    # m or M for Mbps,
    # g or G for Gbps.
    total_bandwidth: "250M"
          
admin_api:  # desc
    # password hash for the admin interface, may be changed via the setting UI.
    password_hash: "$s2$16384$8$1$8zQCf7uWVjbbJ4+HjqTNEzON$dCf/5RdX50464N/JQT6ZJKDZ6VMN74lvHKxw6ooi/YA="

# enable DNS filtering server
dns_filter:
    # where to forward legit requests
    forward_servers:
      - 9.9.9.9
      - 1.1.1.1
    # path to the gravity.db from the Pi Hole project
    blacklist_db: "/opt/vpnhouse/gravity.db"
    # prometheus listen address, disabled if empty
    prom_listen_addr: "localhost:9999"
```

Note that all the necessary configuration options can be provided via the web UI  and does not require the service (or container) restart.

![](https://media.nikonov.tech/config.png)

