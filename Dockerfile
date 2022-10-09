FROM --platform=$BUILDPLATFORM alpine:3.16.2 as alpine

RUN apk add --no-cache ca-certificates tzdata

FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY netlify-dyndns /netlify-dyndns

ENTRYPOINT ["/netlify-dyndns"]
