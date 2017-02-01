test:
	go test -v -race .../gestalt/...

example:
	(cd example && make)

clean:
	(cd example && make clean)

.PHONY: test example clean
