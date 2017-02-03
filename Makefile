test:
	go test -v -race . ./component ./exec ./vars ./test

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
