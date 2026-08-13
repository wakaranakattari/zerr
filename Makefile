GO ?= go

.PHONY: test test-race test-trace bench bench-trace vet fmt check test-zap bench-cmp

test: ## run unit tests
	$(GO) test ./...

test-race: ## run unit tests with the race detector
	$(GO) test -race ./...

test-trace: ## run unit tests with herr_trace stack capture enabled
	$(GO) test -tags herr_trace ./...

bench: ## run benchmarks (production build)
	$(GO) test -bench . -benchmem ./...

bench-trace: ## run benchmarks with stack capture enabled
	$(GO) test -tags herr_trace -bench . -benchmem ./...

test-zap: ## test the zap submodule
	cd zap && $(GO) test ./... && $(GO) vet ./...

bench-cmp: ## head-to-head benchmarks versus oops and cockroachdb/errors
	cd bench && $(GO) test -bench . -benchmem ./...

vet: ## static analysis, both builds
	$(GO) vet ./...
	$(GO) vet -tags herr_trace ./...

fmt: ## check formatting
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

check: fmt vet test test-race test-trace test-zap ## everything