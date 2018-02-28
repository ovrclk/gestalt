test:
	go test -v $$(glide novendor)

example:
	(cd example && make)

deps-install:
	glide install

test-cover:
	export TRAVIS_BRANCH=$${TRAVIS_PULL_REQUEST_BRANCH:-$$TRAVIS_BRANCH}; \
		goveralls -service=travis-pro

coverdeps-install:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls

clean:
	(cd example && make clean)

.PHONY: test example deps-install \
	test-cover coverdeps-install \
	clean
