PACKAGES=$(shell go list ./... | grep -v '/vendor/')
PACKAGES_NOCLITEST=$(shell go list ./... | grep -v '/vendor/' | grep -v github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=${COMMIT_HASH}"

install:
	go install $(BUILD_FLAGS) ./cmd/covenantd
	go install $(BUILD_FLAGS) ./cmd/covenantcli

get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure -v


test: test_unit

test_cli:
	@go test -count 1 -p 1 `go list github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test`

test_unit:
	@go test $(PACKAGES_NOCLITEST)

test_cover:
	@bash tests/test_cover.sh

test_lint:
	gometalinter --disable-all --enable='golint' --vendor ./...

benchmark:
	@go test -bench=. $(PACKAGES_NOCLITEST)
