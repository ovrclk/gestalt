FILES := $(wildcard *.go cmd/*.go)

gestalt: $(FILES)
	go build -o $@

clean:
	rm gestalt || true

.PHONY: clean
