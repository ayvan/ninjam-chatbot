#
#  Makefile for Go
#
APP_NAME=ninjam-chatbot
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_BUILD_RACE=$(GO_CMD) build -race
GO_TEST=$(GO_CMD) test
GO_TEST_VERBOSE=$(GO_CMD) test -v
GO_INSTALL=$(GO_CMD) install -v
GO_CLEAN=$(GO_CMD) clean
GO_DEPS=$(GO_CMD) get -d -v
GO_DEPS_UPDATE=$(GO_CMD) get -d -v -u
GO_FMT=$(GO_CMD) fmt

.PHONY: all build test test-verbose deps clean

all: build
	@echo "Successfully built!";

build: deps
	@echo "Build..."; \
	$(GO_BUILD) -o $(APP_NAME)|| exit 1;

test: build
	@echo "Running tests..."; \
	$(GO_TEST) || exit 1; \
	echo "Tests complete!";

test-verbose: build
	@echo "Running tests verbose..."; \
	$(GO_TEST_VERBOSE) || exit 1; \
	echo "Tests complete!";

deps:
	@echo "Installing dependencies..."; \
	$(GO_DEPS) || exit 1;

clean:
	@rm -rf _vendor || exit 1; \
	rm -f ./$(APP_NAME); \
	echo "Cleared.";