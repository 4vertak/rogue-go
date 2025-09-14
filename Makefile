.PHONY: all build run clean

all: build

build:
	@mkdir build && go build -o build/rogue cmd/rogue/main.go

run:
	@go run cmd/rogue/main.go

clean:
	@rm -f rogue-game statistics.json *.sav go.sum *.json *.json
	@rm -rf build/