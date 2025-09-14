.PHONY: all build run clean

all: build

build:
	@mkdir build && go mod tidy && go build -o build/rogue cmd/rogue/main.go

run:
	@go mod tidy && go run cmd/rogue/main.go

clean:
	@rm -f rogue-game statistics.json *.sav go.sum *.json *.json
	@rm -rf build/