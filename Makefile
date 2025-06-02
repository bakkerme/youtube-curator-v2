.PHONY: run clean-db build test

# Development commands
run:
	go run main.go

build:
	go build -o youtube-curator-v2 ./backend

test:
	go test ./...

clean-db:
	rm -rf ./backend/youtubecurator.db/