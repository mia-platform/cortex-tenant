FROM scratch

COPY --from=golang:1.16 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY cortex-proxy /usr/bin/cortex-proxy
# Use an unprivileged user.
USER 1000

ENTRYPOINT ["/usr/bin/cortex-proxy"]
