GOTEST = go test ./...
GOTEST_WITH_COVERAGE = $(GOTEST) -coverprofile cover.out ./...

.PHONY: test
test:
	$(GOTEST)

.PHONY: cover
cover:
	$(GOTEST_WITH_COVERAGE)