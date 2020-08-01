.PHONY:all
all: lint bin
	go build -v -o ./bin/ ./cmd/...

.PHONY:lint
vet:
	golangci-lint run
	
bin:
	mkdir bin

.PHONY:test
test:
	go test -v ./...

.PHONY:clean
clean:
	rm -v ./bin/*
