TASKS = run build

.PHONY: $(TASKS)

build:
	go build -race -o ./bin/app ./cmd/app/app.go || exit 1

run-urls: build
	./bin/app $(URLS)

run-file: build
	cat ./urls.txt | ./bin/app