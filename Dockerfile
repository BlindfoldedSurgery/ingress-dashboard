FROM golang:1.20-alpine3.18 AS builder

RUN apk update && \
    apk add --no-cache git

WORKDIR $GOPATH/src/torbencarstens/ingress-dashboard
COPY . .

RUN go get -d -v
RUN CGOENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -extldflags=-static" -o /go/bin/ingress-dashboard

FROM scratch

COPY --from=builder /go/bin/ingress-dashboard /go/bin/ingress-dashboard
COPY --from=builder /go/src/torbencarstens/ingress-dashboard/go-templates /go/bin/go-templates

COPY --from=builder /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/go/bin/ingress-dashboard"]
