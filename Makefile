.PHONY: build test fmt vet tidy clean

BIN := bin/buddy
PKG := ./...

build:
	@mkdir -p bin
	go build -trimpath -o $(BIN) ./cmd/buddy

test:
	go test -race -count=1 $(PKG)

fmt:
	gofmt -s -w .

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin/ *.out coverage.txt
