FROM golang:1.18-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/gitlab.com/rarimo/dex-pairs-oracle
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/dex-pairs-oracle /go/src/gitlab.com/rarimo/dex-pairs-oracle


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/dex-pairs-oracle /usr/local/bin/dex-pairs-oracle
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["dex-pairs-oracle"]
