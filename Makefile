test: lint
	go test -cover -count=1 -race ./...

lint:
	golangci-lint run ./...

readme: test
	hype export -o README.md