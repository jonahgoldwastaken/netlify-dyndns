FROM --platform=$BUILDPLATFORM alpine:3.16.2 AS alpine

RUN apk add --no-cache tzdata

FROM scratch

COPY --from=alpine \
	/usr/share/zoneinfo \
	/usr/share/zoneinfo

COPY netlify-dyndns /
ENTRYPOINT ["/netlify-dyndns"]
