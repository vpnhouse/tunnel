
![GHA](https://github.com/vpnhouse/tunnel/actions/workflows/on-merge-into-main.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/vpnhouse/tunnel)](https://goreportcard.com/report/github.com/vpnhouse/tunnel)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/vpnhouse/tunnel?logo=github)
[![Docker pulls](https://img.shields.io/docker/pulls/vpnhouse/tunnel?logo=docker&logoColor=white)](https://hub.docker.com/r/vpnhouse/tunnel)
![GitHub](https://img.shields.io/github/license/vpnhouse/tunnel)


[VPNHouse](https://vpnhouse.net)
==========

A basic, self-contained management service for WireGuard with a self-serve web UI.

- [Quick start](#quick-start)
    - [Server](#server)
    - [Initial setup](#initial-setup)
    - [Add a VPN peer](#add-a-vpn-peer)
    - [Use your new VPN connection](#use-your-new-vpn-connection)
    - [How to update service](#how-to-update-service)
    - [Deep dive](#deep-dive)

### Features

* Self-serve and web based

* QR-Code for convenient mobile client configuration

* Download a client's configuration file

* Zero external dependencies - just a single binary using the wireguard kernel module

* Binary and container deployment


### Requirements

* A host with a kernel that supports WireGuard (all modern kernels).

* A host with [Docker installed &rarr;](https://docs.docker.com/engine/install/ubuntu/#installation-methods).


# Quick start

### Server

Start the server in the Docker container:

```shell
mkdir /opt/vpnhouse-data && # create a directory for the runtime data \
docker run -d \
    --name=vpnhouse-tunnel \
    --restart=always \
    --cap-add NET_ADMIN   `# add extra privilege to manage Wireguard interface` \
    -p 80:80              `# publish web admin port` \
    -p 443:443            `# publish web admin port (SSL)` \
    -p 3000:3000/udp      `# publish Wireguard port` \
    -v /opt/vpnhouse-data/:/opt/vpnhouse/tunnel/   `# mount a host directory with configs` \
    vpnhouse/tunnel:v0.3.4
```

Or, you may use the following [docker-compose](https://raw.githubusercontent.com/vpnhouse/tunnel/main/docs/docker-compose.yaml) file.

Then go to `http://host-ip/` for the initial setup.

### Initial setup

Set the password and the network subnet for VPN clients:

<img src="https://media.nikonov.tech/initial-with-ssl2.png" style="width: 60%; max-width: 240px" alt="Peers" />

Tick **I have a domain name** only if you have a domain, as well as a DNS record that points to *this* machine.

If you tick the **Issue SSL certificate** we will automatically obtain and renew the valid SSL certificate via LetsEncrypt.

**Reverse proxy**: chose this option if you have the webserver configured on this machine,
and you want to use it as a reverse proxy for the VPNHouse service.


### Add a VPN peer

Click "Add new" to create a connection to your new VPN server.

Give it a name and optional expiration date. The suggested IP address is perfectly valid and can be used. 
You don’t have to change it. But if you have a sound reason, you may activate the field and set the IP address by hand.

<img src="https://media.nikonov.tech/add_peer.png" style="width: 60%; max-width: 240px" alt="Peers" />


### Use your new VPN connection

1. [Download &rarr;](https://www.wireguard.com/install/) the official WireGuard client for your OS/device.

2. Use the QR-code to set-up your mobile client, [or follow our step-by-step guide](https://github.com/vpnhouse/tunnel/blob/main/docs/mobile.md).


<img src="https://media.nikonov.tech/config-qr.png" style="width: 60%; max-width: 240px" alt="QR" />


3. The "Show config" button shows the configuration in the text format. Use it for the desktop client, [or follow our step-by-step guide](https://github.com/vpnhouse/tunnel/blob/main/docs/desktop.md).

<img src="https://media.nikonov.tech/config-text.png" style="width: 60%; max-width: 240px" alt="QR" />

### How to update service

If you're using our [docker-compose](https://raw.githubusercontent.com/vpnhouse/tunnel/main/docs/docker-compose.yaml) file, just change container's version inside and run:

```shell
docker-compose up -d
```

If you started service by `docker run` command as recommended in [server](#server) quick start section, then just stop and remove old container and then start a new one:

```shell
docker stop vpnhouse-tunnel; docker rm -f vpnhouse-tunnel && \
docker run -d \
    --name=vpnhouse-tunnel \
    --restart=always \
    --cap-add NET_ADMIN   `# add extra privilege to manage Wireguard interface` \
    -p 80:80              `# publish web admin port` \
    -p 443:443            `# publish web admin port (SSL)` \
    -p 3000:3000/udp      `# publish Wireguard port` \
    -v /opt/vpnhouse-data/:/opt/vpnhouse/tunnel/   `# mount a host directory with configs` \
    vpnhouse/tunnel:v0.3.5
```

### Deep dive

* [Configuration file reference](https://github.com/vpnhouse/tunnel/blob/main/docs/config.md)

* [Building it locally](https://github.com/vpnhouse/tunnel/blob/main/docs/building.md)

