PACKAGE=thirtyeightybot

.PHONY: test run build

test:
	go test

run:
	go run main.go

build:
	go build -o $(PACKAGE) main.go