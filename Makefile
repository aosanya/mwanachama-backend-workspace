.PHONY: build test test-pg vet clean

## Verify the module compiles cleanly.
build:
	go build ./...

## Unit tests (no DB required).
test:
	go test ./...

## Integration tests against a real Postgres instance.
## Expects POSTGRES_URL; the index behaviour is only covered here.
test-pg:
	go test -tags=integration ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
