all: bin
	go build -v -o ./bin/  ./cmd/...
	
bin:
	mkdir bin

clean:
	rm -v ./bin/*
