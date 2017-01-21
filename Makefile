test:
	go test -v -race ./test ./vars

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
