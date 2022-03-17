#!/bin/sh

mkdir -p /opt/vpnhouse/tunnel/trusted_rsa
mkdir -p /opt/vpnhouse/tunnel/config

iptables -A POSTROUTING -t nat -j MASQUERADE

exec /usr/local/bin/tunnel-node
