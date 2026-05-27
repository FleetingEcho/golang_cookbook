.PHONY: run build test seed clean

run:
	go run ./

build:
	go build -o bin/issue_tracker ./

test:
	go test -v -count=1 ./...

seed:
	@mkdir -p data uploads
	@rm -f data/issue_tracker.db
	@sqlite3 data/issue_tracker.db < migrations/0001_init.sql
	@sqlite3 data/issue_tracker.db < scripts/seed.sql
	@echo "seeded data/issue_tracker.db"

clean:
	rm -rf bin/ data/issue_tracker.db
	@find uploads -type f ! -name '.gitkeep' -delete 2>/dev/null || true
