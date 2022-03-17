#!/bin/sh

iptables -A POSTROUTING -t nat -j MASQUERADE

exec /usr/local/bin/tunnel-node
