.PHONY: run clean-db

run:
	go run main.go

clean-db:
	rm -rf youtubecurator.db/ 