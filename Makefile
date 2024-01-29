GO_FILES=`go list ./... | grep -v -E "mock|store|test|fake|cmd"`

.PHONY: ci-test
ci-test:
	@go test $(GO_FILES) -coverprofile .cover.txt
	@go tool cover -func .cover.txt
	@rm .cover.txt

.PHONY: ci-build
ci-build:
	rm -rf bin
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/hutao -ldflags "-X hutao/pod.appRelease=${release}" main.go
	cp /usr/local/go/lib/time/zoneinfo.zip bin/zoneinfo.zip

.PHONY: build
build:
	rm -rf bin
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/hutao -ldflags "-X github.com/zealerFT/https-certificate-check/pod.appRelease=${release}" main.go

.PHONY: boot
boot:
	docker-compose up --no-recreate -d

.PHONY: clean
clean:
	docker-compose down --remove-orphans
	docker rm -f $(docker ps -a | grep Exit | awk '{ print $1 }') || true
