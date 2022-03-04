all: build

TAG = v0.3.0

build: clean deps fmt
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kube-applier

container:
	docker build -t kube-applier:$(TAG) .

clean:
	rm -f kube-applier

fmt:
	go fmt

vet:
	go vet

test-unit: clean deps fmt build vet
	go test -v --race ./...

.PHONY: all deps build container clean fmt test-unit vet
