FROM scratch

ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/dns-mokka"]
COPY  $TARGETPLATFORM/dns-mokka /usr/bin/dns-mokka
