FROM golang:1.10-alpine3.7

RUN apk add --no-cache git

WORKDIR /go/src/github.com/hypnoglow/kubernetes-cookbook/grpc-load-balancing/dns-resolver

COPY . .

RUN cd ./client \
    && go get -v ./... \
    && go build -o client

FROM alpine:3.7

RUN apk add --no-cache ca-certificates

COPY --from=0 /go/src/github.com/hypnoglow/kubernetes-cookbook/grpc-load-balancing/dns-resolver/client/client /app/client

ENTRYPOINT ["/app/client"]
