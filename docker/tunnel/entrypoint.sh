#!/bin/sh

mkdir -p /opt/uranium/tunnel/trusted_rsa
mkdir -p /opt/uranium/tunnel/config

iptables -A POSTROUTING -t nat -j MASQUERADE

exec /usr/local/bin/tunnel-node
