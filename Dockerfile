FROM debian:wheezy

WORKDIR /data
ENTRYPOINT ["/opt/bin/shadowtunnel"]
ADD shadowtunnel /opt/bin/shadowtunnel

USER nobody

