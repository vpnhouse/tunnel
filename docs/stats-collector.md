Stats collector container
========================

Running VPNHouse stats collector service in docker:

```
mkdir /opt/vh-stats-data
docker run -d \  
    --restart=always \
    --name=vpnhouse-stats \
    -p 127.0.0.1:8123:8123 \
    -v /opt/vh-stats-data:/extstat-data/ \
    -e VPNHOUSE_EXTSTAT_USERNAME=username \
    -e VPNHOUSE_EXTSTAT_PASSWORD=secret\
    vpnhouse/statserver:latest
```

Username and password are used for the web auth.
