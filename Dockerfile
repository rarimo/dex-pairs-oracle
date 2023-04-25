FROM golang:1.19-alpine as buildbase

RUN apk add build-base git

ARG CI_JOB_TOKEN

WORKDIR /go/src/gitlab.com/rarimo/dex-pairs-oracle

COPY . .

ENV GO111MODULE="on"
ENV CGO_ENABLED=1
ENV GOOS="linux"

RUN echo "machine gitlab.com login gitlab-ci-token password $CI_JOB_TOKEN" > ~/.netrc
RUN git config --global url."https://gitlab-ci-token:$CI_JOB_TOKEN@gitlab.com/".insteadOf https://gitlab.com/
RUN go mod tidy
RUN go mod vendor
RUN go build -o /usr/local/bin/dex-pairs-oracle gitlab.com/rarimo/dex-pairs-oracle

###

FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/dex-pairs-oracle /usr/local/bin/dex-pairs-oracle
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["dex-pairs-oracle"]
