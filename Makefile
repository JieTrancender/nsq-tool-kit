# Build all by default, even if it's not first
.DEFAULT_GOAL := all

.PHONY: all
all: build

ROOT_PACKAGE=github.com/JieTrancender/nsq-tool-kit
VERSION_PACKAGE=github.com/marmotedu/component-base/pkg/version

include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk

### build:             Show Makefile rules
.PHONY: build
build:
	@$(MAKE) go.build

APPS = nsq-tool-kit

$(BUILDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BUILDFLAGS} -o $@ ./

$(APPS): %: $(BUILDDIR)/%

### help:             Show Makefile rules
.PHONY: help
help:
	@echo Makefile rules:
	@echo
	@grep -E '^### [-A-Za-z0-9_]+:' Makefile | sed 's/###/   /'

### test:             Exec go test
.PHONY: test
test:
	@$(MAKE) go.test

### lint:             Exec lint
.PHONY: lint
lint:
	golangci-lint cache clean
	golangci-lint run --tests=false ./...

### clean:            Clean up generated files
clean:
	rm -rf $(BUILDDIR)

### tidy:             Exec go mod tidy
tidy:
	go mod tidy
