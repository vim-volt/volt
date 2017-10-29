
NAME := volt
SRC := $(wildcard *.go */*.go)
VERSION := $(shell git describe --tags)
DEVEL_LDFLAGS := -X github.com/vim-volt/go-volt/cmd.version=$(VERSION)
RELEASE_LDFLAGS := $(DEVEL_LDFLAGS) -extldflags '-static'
RELEASE_OS := linux windows darwin
RELEASE_ARCH := amd64 386

DIST_DIR := dist
BIN_DIR := bin

all: $(BIN_DIR)/$(NAME)

$(BIN_DIR)/$(NAME): $(SRC)
	go build -ldflags "$(DEVEL_LDFLAGS)" -o $(BIN_DIR)/$(NAME)

setup:
	@which go >/dev/null 2>&1   || (echo '[Error] You need to install go,make commands'; exit 1)
	@which make >/dev/null 2>&1 || (echo '[Error] You need to install go,make commands'; exit 1)
	go get github.com/golang/dep/cmd/dep
	dep ensure

precompile:
	go build -a -i -o $(BIN_DIR)/$(NAME)
	rm $(BIN_DIR)/$(NAME)

# Make static-linked binaries and tarballs
release: $(SRC)
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


.PHONY: all setup precompile release
