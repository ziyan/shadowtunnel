FROM busybox

ADD shadowtunnel /bin/shadowtunnel

USER nobody

ENTRYPOINT ["/bin/shadowtunnel"]

