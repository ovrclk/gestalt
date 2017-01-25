test:
	go test -v -race ./...

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
