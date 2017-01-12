FILES := $(wildcard *.go ../*.go)

gestalt: $(FILES)
	go build -o $@

clean:
	rm gestalt || true

.PHONY: clean
