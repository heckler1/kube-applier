all: build

TAG = v0.3.0

build: clean deps fmt generate
	GOOS=linux GOARCH=amd64 go build -o kube-applier

deps:
	go mod tidy
	go mod vendor

container:
	docker build -t kube-applier:$(TAG) .

clean:
	rm -f kube-applier

fmt:
	go fmt

vet:
	go vet

lint:
	golangci-lint run ./...

test: clean deps fmt vet build
	go test -v --race ./...

.PHONY: all deps build container clean fmt test-unit vet lint generate

bin/mockgen: $(shell find ./vendor/github.com/golang/mock -name '*.go')
	go build -o $@ ./vendor/github.com/golang/mock/mockgen

generate: git/mock_gitutil.go kube/mock_client.go run/mock_batch_applier.go sysutil/mock_clock.go sysutil/mock_filesystem.go

git/mock_gitutil.go: git/types.go bin/mockgen
	./bin/mockgen  -source=$< -package=git -destination=$@.new
	mv $@.new $@

kube/mock_client.go: kube/client.go bin/mockgen
	./bin/mockgen  -source=$< -package=kube -destination=$@.new
	mv $@.new $@

run/mock_batch_applier.go: run/batch_applier.go bin/mockgen
	./bin/mockgen  -source=$< -package=run -destination=$@.new
	mv $@.new $@

sysutil/mock_clock.go: sysutil/clock.go bin/mockgen
	./bin/mockgen  -source=$< -package=sysutil -destination=$@.new
	mv $@.new $@

sysutil/mock_filesystem.go: sysutil/filesystem.go bin/mockgen
	./bin/mockgen  -source=$< -package=sysutil -destination=$@.new
	mv $@.new $@