FROM alpine:3.16

LABEL org.opencontainers.image.source="https://github.com/0xERR0R/dns-mokka" \
      org.opencontainers.image.url="https://github.com/0xERR0R/dns-mokka" \
      org.opencontainers.image.source=https://github.com/0xERR0R/dns-mokka \
      org.opencontainers.image.title="simple DNS mocker"
      

ENTRYPOINT ["/usr/bin/dns-mokka"]
COPY dns-mokka /usr/bin/dns-mokka