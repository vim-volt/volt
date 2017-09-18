
NAME := volt
SRC := $(wildcard *.go */*.go)
VERSION := $(shell git describe --tags)
DEVEL_LDFLAGS := -X main.version=$(VERSION)
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
	go get github.com/Masterminds/glide
	glide install
	make precompile

precompile:
	go build -i -o $(BIN_DIR)/$(NAME)
	rm $(BIN_DIR)/$(NAME)

# Make static-linked binaries and tarballs
release: $(SRC)
	make npm_build_prod
	@for os in $(RELEASE_OS); do \
		for arch in $(RELEASE_ARCH); do \
			if [ $$os = windows ]; then \
				echo "Making zip for ... os=$$os, arch=$$arch"; \
				GOOS=$$os GOARCH=$$arch go build -tags netgo -installsuffix netgo -ldflags "$(RELEASE_LDFLAGS)" -o $(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch/$(NAME).exe; \
				(cd $(DIST_DIR) && zip -qr $(NAME)-$(VERSION)-$$os-$$arch.zip $(NAME)-$(VERSION)-$$os-$$arch); \
			else \
				echo "Making tarball for ... os=$$os, arch=$$arch"; \
				GOOS=$$os GOARCH=$$arch go build -tags netgo -installsuffix netgo -ldflags "$(RELEASE_LDFLAGS)" -o $(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch/$(NAME); \
				strip $(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch/$(NAME) 2>/dev/null; \
				(cd $(DIST_DIR) && tar czf $(NAME)-$(VERSION)-$$os-$$arch.tar.gz $(NAME)-$(VERSION)-$$os-$$arch); \
			fi; \
			rm -r $(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch; \
		done; \
	done


.PHONY: all setup precompile release
