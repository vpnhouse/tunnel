# Configuration file reference

Defualt path is `/opt/uranium/tunnel/static.yaml`.

```yaml
# static.yaml
log_level: debug
sqlite_path: /opt/uranium/tunnel/db.sqlite3
http_listen_addr: :8085

# enable the documentation server under the `/rapidoc/` path
rapidoc: true

wireguard:
    # interface name, the interface will be allocated automatically
    interface: "uwg0"
    # public IPv4 of the server, announced to clients
    server_ipv4: "1.2.3.4"
    # wireguard port, announced to clients
    server_port: 3000
    # keepalive interval
    keepalive: 60
    # subnet for VPN clients, server will take the first available address automatically
    subnet: "10.235.0.0/24"
    # a list of DNS servers to announce to clients
    dns:
        - 8.8.8.8
        - 8.8.4.4

ssl: # optinal SSL & LetsEncrypt configuration
    domain: "the-machine.example.com"  # DNS name to issue certificate with
    listen_addr: ":443"                # listen address for HTTPS server
    dir: "/opt/uranium/tunnel/"        # directory to cache SSL certificates data
```

Note that all the necessary configuration options can be provided via the web UI  and does not require the service (or container) restart.

![](https://media.nikonov.tech/config.png)

