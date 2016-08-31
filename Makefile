install:
	go install -v

build:
	go build -v ./...

lint:
	$(GOPATH)/bin/golint ./...
	go vet ./...

test:
	go test -v ./...

cover:
	go test -v ./... --cover

deps: dev-deps
	go get github.com/nats-io/nats
	go get gopkg.in/redis.v3
	go get github.com/ernestio/ernest-config-client
	go get github.com/ernestio/builder-library

dev-deps:
	go get github.com/golang/lint/golint

clean:
	go clean
	rm -f gpb-firewalls-microservice

dist-clean:
	rm -rf pkg src bin

ci-deps:
