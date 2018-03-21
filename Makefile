test:
	go test $$(glide novendor)

example:
	(cd example && make)

deps-install:
	glide install

test-cover:
	go test -covermode=count -coverprofile=test.cov ./...
	goveralls -coverprofile=test.cov -service=travis-pro

coverdeps-install:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls

clean:
	(cd example && make clean)

.PHONY: test example deps-install \
	test-cover coverdeps-install \
	clean
