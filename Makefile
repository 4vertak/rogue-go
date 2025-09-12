.PHONY: all build run clean

all: build

build:
	go build -o rogue-game main.go structure.go logic.go interface.go

run: build
	./rogue-game

clean:
	rm -f rogue-game statistics.json *.sav go.mod go.sum *.json