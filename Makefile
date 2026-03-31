.PHONY: redact server test clean

redact:
	@mkdir -p bin
	go build -o bin/redact ./cmd/redact

server:
	@mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin
