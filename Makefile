test:
	go test ./...

example:
	(cd example && make)

test-cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	curl -s https://codecov.io/bash | bash

clean:
	(cd example && make clean)

.PHONY: test example deps-install \
	test-cover coverdeps-install \
	clean
