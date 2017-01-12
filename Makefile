FILES := $(wildcard *.go cmd/*.go)

cmd/gestalt: $(FILES)
	cd cmd && make

test:
	go test -v ./test

clean:
	cd cmd && make clean

.PHONY: test clean
