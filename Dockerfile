FROM scratch

COPY netlify-dyndns /netlify-dyndns

ENTRYPOINT ["/netlify-dyndns"]
