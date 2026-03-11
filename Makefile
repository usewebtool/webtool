.PHONY: build test test-integration clean

build:
	go build -o dist/webtool .

test:
	go test ./...

test-integration: build
	go test -tags integration ./test/integration/ -v -count=1

clean:
	rm -rf dist/webtool
