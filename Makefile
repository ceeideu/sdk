LINT=golangci-lint

clean:
	rm -f $(COVERAGE_OUT)

lint:
	./$(LINT) run

test:
	$(GOTEST) -coverprofile=$(COVERAGE_OUT) ./...
