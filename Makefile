.PHONY:all
all: vet bin
	go build -v -o ./bin/  ./cmd/...

.PHONY:vet
vet:
	go vet ./...
	
bin:
	mkdir bin

.PHONY:test
test:
	go test -v ./...

.PHONY:clean
clean:
	rm -v ./bin/*
