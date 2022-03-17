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
  listen_addr: :80

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
          
admin_api:  # desc
    # login for the admin interface, may be changed via the setting UI.
    user_name: admin
    # password hash for the admin interface, may be changed via the setting UI.
    password_hash: $s2$16384$8$1$8zQCf7uWVjbbJ4+HjqTNEzON$dCf/5RdX50464N/JQT6ZJKDZ6VMN74lvHKxw6ooi/YA=

ssl: # optional SSL & LetsEncrypt configuration
    domain: "the-machine.example.com"  # DNS name to issue certificate with
    listen_addr: ":443"                # listen address for HTTPS server
    dir: "/opt/vpnhouse/tunnel/"        # directory to cache SSL certificates data
```

Note that all the necessary configuration options can be provided via the web UI  and does not require the service (or container) restart.

![](https://media.nikonov.tech/config.png)

