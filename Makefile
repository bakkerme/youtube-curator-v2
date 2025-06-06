.PHONY: run clean-db build test

# Development commands
run-backend:
	cd backend && go run main.go

run-backend-air:
	cd backend && air .

run-frontend:
	cd frontend && npm run dev

build:
	go build -o youtube-curator-v2 ./backend

test:
	go test ./...

clean-db:
	rm -rf ./backend/youtubecurator.db/