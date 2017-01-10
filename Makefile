FILES := $(wildcard *.go cmd/*.go)

cmd/gestalt: $(FILES)
	cd cmd && make

clean:
	cd cmd && make clean

.PHONY: clean
