Uranium VPN
===========

A basic, self-contained management service for WireGuard with a self-serve web UI.

- [Quick start](#quick-start)
    - [Server](#server)
    - [Initial setup](#initial-setup)
    - [Add a VPN peer](#add-a-vpn-peer)
    - [Use your new VPN connection](#use-your-new-vpn-connection)
    - [Deep dive](#deep-dive)

### Features

* Self-serve and web based

* QR-Code for convenient mobile client configuration

* Download a client's configuration file

* Zero external dependencies - just a single binary using the wireguard kernel module

* Binary and container deployment


### Requirements

* A host with a kernel that supports WireGuard (all modern kernels).

* A host with [Docker installed &rarr;](https://docs.docker.com/get-docker/).


# Quick start

### Server

Start the server in the Docker container:

```shell
$ mkdir uranium-data # create a directory for the runtime data
$ docker run -d \
    --name=uranium-tunnel \
    --restart=always \
    --cap-add NET_ADMIN \  # add extra privilege to manage Wireguard interface
    -p 80:80 \             # publish web admin port
    -p 443:443 \           # publish web admin port (SSL)
    -p 3000:3000/udp \     # publish Wireguard port
    -v uranium-data:/opt/uranium/tunnel/ \  # mount a host directory with configs
    codenameuranium/tunnel:latest-personal
```

Or, you may use the following [docker-compose](https://gist.github.com/835d4ac1b3c2a203cd53f5d9fb5e7ab8) file.

Then go to `http://host-ip/` for the initial setup.

### Initial setup

Provide an initial configuration using the simple configuration form:

<img src="https://media.nikonov.tech/initial-with-ssl.png" style="width: 60%; max-width: 240px" alt="Peers" />

You may also specify the domain name if you have one. We'll automatically issue
the valid SSL certificate via LetsEncrypt.

You will be redirected to the login form. Use your password and the `admin` as the user name to log into the peer management console.


### Add a VPN peer

Click "Add new" to create a connection to your new VPN server.

Give it a name and optional expiration date. Also, you may change the IP address,
but the one suggested by the creation form is perfectly valid and can be used.

<img src="https://media.nikonov.tech/add-peer-form.png" style="width: 60%; max-width: 240px" alt="Peers" />


### Use your new VPN connection

1. [Download &rarr;](https://www.wireguard.com/install/) the official WireGuard client for your OS/device.

2. Use the QR-code to set-up your mobile client, [or follow our step-by-step guide](https://github.com/Codename-Uranium/tunnel/blob/main/docs/mobile.md).


<img src="https://media.nikonov.tech/config-qr.png" style="width: 60%; max-width: 240px" alt="QR" />


3. The "Show config" button shows the configuration in the text format. Use it for the desktop client, [or follow our step-by-step guide](https://github.com/Codename-Uranium/tunnel/blob/main/docs/desktop.md).

<img src="https://media.nikonov.tech/config-text.png" style="width: 60%; max-width: 240px" alt="QR" />


### Deep dive

* [Configuration file reference](https://github.com/Codename-Uranium/tunnel/blob/main/docs/config.md)

* [Building it locally](https://github.com/Codename-Uranium/tunnel/blob/main/docs/building.md)
