FROM golang:alpine AS prepare
RUN apk update && apk add --no-cache ca-certificates git && update-ca-certificates

FROM scratch
COPY --from=prepare /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=prepare /usr/local/bin/ /usr/local/bin/
WORKDIR /app
COPY artifacts .
ENTRYPOINT ["./service"]
