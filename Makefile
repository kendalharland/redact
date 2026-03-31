.PHONY: redact test clean

redact:
	@mkdir -p bin
	go build -o bin/redact ./cmd/redact

test:
	go test ./...

clean:
	rm -rf bin
