TASKS = run-urls run-file build deps

.PHONY: $(TASKS)

deps:
	go mod tidy -compat=1.17
	go mod vendor

build: deps
	go build -race -o ./bin/app ./cmd/app/app.go || exit 1

run-urls: build
	MAX_CONCURRENCY=2  ./bin/app $(URLS)

run-file: build
	cat ./urls.txt | MAX_CONCURRENCY=10 ./bin/app