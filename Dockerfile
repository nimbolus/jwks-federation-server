FROM docker.io/golang:1.23.2-alpine as builder
RUN apk --no-cache add ca-certificates

COPY . /go/srv/github.com/jwks-federation-server
WORKDIR /go/srv/github.com/jwks-federation-server
RUN CGO_ENABLED=0 go build -o /jwks-federation-server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /jwks-federation-server /jwks-federation-server

ENTRYPOINT ["/jwks-federation-server"]
