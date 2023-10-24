FROM golang:1.20-alpine as buildbase

RUN apk add build-base git

WORKDIR /go/src/github.com/rarimo/dex-pairs-oracle

COPY . .

ENV GO111MODULE="on"
ENV CGO_ENABLED=1
ENV GOOS="linux"

RUN go mod tidy
RUN go mod vendor
RUN go build -o /usr/local/bin/dex-pairs-oracle github.com/rarimo/dex-pairs-oracle

###

FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/dex-pairs-oracle /usr/local/bin/dex-pairs-oracle
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["dex-pairs-oracle"]
