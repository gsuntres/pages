.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

test-all: ## Run all tests
	RUN_INTEGRATION=yes \
	go run gotest.tools/gotestsum@latest \
	-- -failfast \
	-race ./...

test-watch: ## Watch tests
	RUN_INTEGRATION=yes \
	go run gotest.tools/gotestsum@latest \
	--watch \
	-- \
	-failfast \
	-race \
	./...

.PHONY: run

