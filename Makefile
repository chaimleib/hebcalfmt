test:
	go test -race -timeout 15s \
		-coverprofile coverage.out \
		./...
