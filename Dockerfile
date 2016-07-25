FROM golang:1.6.2-alpine

RUN apk add --update git && apk add --update make && rm -rf /var/cache/apk/*

ADD . /go/src/github.com/ernestio/router-builder
WORKDIR /go/src/github.com/ernestio/router-builder

RUN make deps && make install

ENTRYPOINT ./entrypoint.sh
