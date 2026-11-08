.PHONY: all build run clean

GO ?= go

all: build

build:
	mkdir -p build
	$(GO) build -mod=readonly -o build/rogue ./cmd/rogue

run: build
	./build/rogue

clean:
	rm -f build/rogue
