all: bin
	go build -v -o ./bin/ 
	
bin:
	mkdir bin

clean:
	rm -v ./bin/*
