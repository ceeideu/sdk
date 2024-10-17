LINT=golangci-lint

lint:
	./$(LINT) run

test:
	go test ./...
