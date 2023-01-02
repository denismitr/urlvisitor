TASKS = run build

.PHONY: $(TASKS)

build:
	go build -race -o ./bin/app ./cmd/app/app.go || exit 1

run: build
	./bin/app $(URLS)