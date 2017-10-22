.PHONY: test deps

test:
	GO_ROUTER_ENABLE_LOGGING=1 go test -v $(glide novendor)

deps:
	dep ensure
