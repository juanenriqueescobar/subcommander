
-include .env

VERSION=v0.1.1

BUILD_OS=linux # darwin windows
BUILD_ARCH=amd64 arm64
BUILD_OA=$(foreach os,$(BUILD_OS),$(addprefix $(os).,$(BUILD_ARCH)))
BUILD=$(addprefix build.,$(BUILD_OA))
PUBLISH=$(addprefix publish.,$(BUILD_OA))

ARCH=$(subst .,,$(suffix $@))
OS=$(subst .,,$(suffix $(basename $@)))

$(BUILD):
	@echo "building $(OS) $(ARCH)"
	@mkdir -p bin	
	@GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="-w -s -X main.Version=$(VERSION)" -o bin/subcommander-$(OS)-$(ARCH)-$(VERSION) cmd/main.go
	@cd bin && sha256sum subcommander-$(OS)-$(ARCH)-$(VERSION) > subcommander-$(OS)-$(ARCH)-$(VERSION).sha256sum

build: $(BUILD)

# publish: upload to s3 the same version in both archs
# run with infra-prod user
publish: $(PUBLISH)

$(PUBLISH):
	@echo "publish $(OS) $(ARCH)"
	@echo "upload binaries/subcommander/subcommander-$(OS)-$(ARCH)-$(VERSION)"
	@aws s3api put-object \
		--cache-control max-age=31536000 \
		--bucket $(BUCKET) \
		--key binaries/subcommander/subcommander-$(OS)-$(ARCH)-$(VERSION)\
		--body bin/subcommander-$(OS)-$(ARCH)-$(VERSION)
	@echo "upload binaries/subcommander/subcommander-$(OS)-$(ARCH)-$(VERSION).sha256sum"
	@aws s3api put-object \
		--cache-control max-age=31536000 \
		--bucket $(BUCKET) \
		--key binaries/subcommander/subcommander-$(OS)-$(ARCH)-$(VERSION).sha256sum \
		--body bin/subcommander-$(OS)-$(ARCH)-$(VERSION).sha256sum
