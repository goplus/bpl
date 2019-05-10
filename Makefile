all:
	go install -v ./...

rebuild:
	go install -a -v ./...

install: all
	@mkdir -p ~/.qbpl/formats
	cp formats/*.bpl ~/.qbpl/formats/

test:
	go test ./...

testv:
	go test -v ./...

clean:
	go clean -i ./...

fmt:
	gofmt -w=true ./
