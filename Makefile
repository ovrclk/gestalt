test:
	go test -race . ./component ./exec ./vars

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
