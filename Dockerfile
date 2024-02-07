FROM golang:1.22-alpine3.18 AS builder

RUN apk update && \
    apk add --no-cache git

WORKDIR $GOPATH/src/torbencarstens/ingress-dashboard
COPY dashboard/ dashboard/
COPY utils/ utils/
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go

RUN go get -d -v
RUN CGOENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -extldflags=-static" -o /go/bin/ingress-dashboard

FROM scratch

COPY --from=builder /go/bin/ingress-dashboard /go/bin/ingress-dashboard

COPY --from=builder /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY go-templates/ /go-templates/
COPY public/ /public/
ENV GIN_MODE=release

ENTRYPOINT ["/go/bin/ingress-dashboard"]
