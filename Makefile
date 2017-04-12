test:
	go test -race . ./component ./exec ./vars ./util

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
