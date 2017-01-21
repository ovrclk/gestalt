FILES := $(wildcard *.go ../*.go ../*/*.go)

OBJECT := example

$(OBJECT): $(FILES)
	go build -o $@

clean:
	rm $(OBJECT) || true

.PHONY: $(OBJECT)
