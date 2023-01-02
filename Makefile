TASKS = run-urls run-file build deps build-race

.PHONY: $(TASKS)

deps:
	go mod tidy
	go mod vendor

build-race: deps
	go build -race -o ./bin/app ./cmd/app/app.go || exit 1

build: deps
	go build -o ./bin/app ./cmd/app/app.go || exit 1

run-urls: build
	MAX_CONCURRENCY=2  ./bin/app $(URLS)

run-file: build
	cat ./urls.txt | MAX_CONCURRENCY=10 ./bin/app