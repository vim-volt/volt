
NAME := volt
SRC := $(wildcard *.go */*.go)
VERSION := $(shell sed -n -E 's/var version string = "([^"]+)"/\1/p' cmd/version.go)
RELEASE_LDFLAGS := -extldflags '-static'
RELEASE_OS := linux windows darwin
RELEASE_ARCH := amd64 386

DIST_DIR := dist
BIN_DIR := bin

all: $(BIN_DIR)/$(NAME)

$(BIN_DIR)/$(NAME): $(SRC)
	go build -o $(BIN_DIR)/$(NAME)

precompile:
	go build -a -i -o $(BIN_DIR)/$(NAME)
	rm $(BIN_DIR)/$(NAME)

test:
	make
	go test -v -race ./...

# Make static-linked binaries and tarballs
release: $(BIN_DIR)/$(NAME)
	rm -fr $(DIST_DIR)
	@for os in $(RELEASE_OS); do \
		for arch in $(RELEASE_ARCH); do \
			if [ $$os = windows ]; then \
				exe=$(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch.exe; \
				echo "Creating $$exe ... (os=$$os, arch=$$arch)"; \
				GOOS=$$os GOARCH=$$arch go build -tags netgo -installsuffix netgo -ldflags "$(RELEASE_LDFLAGS)" -o $$exe; \
			else \
				exe=$(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch; \
				echo "Creating $$exe ... (os=$$os, arch=$$arch)"; \
				GOOS=$$os GOARCH=$$arch go build -tags netgo -installsuffix netgo -ldflags "$(RELEASE_LDFLAGS)" -o $$exe; \
				strip $(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch 2>/dev/null; \
				true; \
			fi; \
		done; \
	done


.PHONY: all precompile test release
