.PHONY: run clean-db build docker-build docker-up docker-down docker-logs validate test

# Development commands
run:
	go run main.go

build:
	go build -o youtube-curator-v2 .

test:
	go test ./...

clean-db:
	rm -rf youtubecurator.db/

clean: clean-db
	rm -f youtube-curator-v2
	docker compose down || true
	docker rmi youtube-curator-v2 || true 