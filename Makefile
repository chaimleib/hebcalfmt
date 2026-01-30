test:
	go test -race -timeout 5s \
		-coverprofile coverage.out \
		./...

.PHONY: test
