# Building

### Static binary

Requirements:

* Linux host

* Go >=1.17

* gcc >=9

* musl-dev >=1.2


To build the executable, simply run:

```shell
make build
```

or, using the Go compiler directly:

```shell
go build -o tunnel-node cmd/tunnel/main.go
```

Then run it with the configuration directory specified:

```shell
./tunnel-node -cfg vpnhouse-data
```

Note: you may have to use `sudo` since the tunnel needs an access to the
netlink to be able to create and manage the Wireguard network interface.


### Docker image


To build the docker image, define the `DOCKER_IMAGE` env variable
with your Docker registry address, username, image name and version:

```shell
DOCKER_IMAGE=username/tunnel:version make docker/build/personal
```

