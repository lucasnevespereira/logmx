VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run clean lint release-patch release-minor release-major

build:
	@go build -ldflags "$(LDFLAGS)" -o bin/logmx ./cmd/logmx

run:
	@go run ./cmd/logmx

clean:
	@rm -rf bin

lint:
	@go vet ./...
	@golangci-lint run

# --- releases ---

LATEST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
NEXT_PATCH := $(shell echo $(LATEST_TAG) | awk -F. '{printf "%s.%s.%d", $$1, $$2, $$3+1}')
NEXT_MINOR := $(shell echo $(LATEST_TAG) | awk -F. '{printf "%s.%d.0", $$1, $$2+1}')
NEXT_MAJOR := $(shell echo $(LATEST_TAG) | awk -F. '{split($$1,a,"v"); printf "v%d.0.0", a[2]+1}')

define tag_and_push
	@echo "$(LATEST_TAG) -> $(1)"
	@printf "Push $(1)? [y/N] " && read ans && [ "$$ans" = "y" ] || (echo "aborted"; exit 1)
	@git tag $(1) && git push origin $(1)
	@echo "done — release workflow started"
endef

release-patch:
	$(call tag_and_push,$(NEXT_PATCH))

release-minor:
	$(call tag_and_push,$(NEXT_MINOR))

release-major:
	$(call tag_and_push,$(NEXT_MAJOR))
