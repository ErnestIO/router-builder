FROM golang:1.7.1-alpine

RUN apk add --update git && apk add --update make && rm -rf /var/cache/apk/*

ADD . /go/src/github.com/${GITHUB_ORG:-ernestio}/router-builder
WORKDIR /go/src/github.com/${GITHUB_ORG:-ernestio}/router-builder

RUN make deps && make install

ENTRYPOINT ./entrypoint.sh
