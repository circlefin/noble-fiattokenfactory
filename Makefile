.PHONY: proto-all proto-format proto-lint proto-gen format heighliner test-e2e test-unit test build install

all: proto-all format lint test-unit build

###############################################################################
###                                  Build                                  ###
###############################################################################

build:
	@echo "🤖 Building simd..."
	@cd simapp && make build
	@echo "✅ Completed build!"

install:
	@echo "🤖 Installing simd..."
	@cd simapp && make install
	@echo "✅ Completed install!"

###############################################################################
###                          Formatting & Linting                           ###
###############################################################################

gofumpt_cmd=mvdan.cc/gofumpt

format:
	@echo "🤖 Running formatter..."
	@go run $(gofumpt_cmd) -l -w .
	@echo "✅ Completed formatting!"

###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.34.0
BUILDER_VERSION=0.14.0

proto-all: proto-format proto-lint proto-gen

proto-format:
	@echo "🤖 Running protobuf formatter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "✅ Completed protobuf formatting!"

proto-gen:
	@echo "🤖 Generating code from protobuf..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./proto/generate.sh
	@echo "✅ Completed code generation!"

proto-lint:
	@echo "🤖 Running protobuf linter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "✅ Completed protobuf linting!"

###############################################################################
###                                 Testing                                 ###
###############################################################################

heighliner:
	@echo "🤖 Building image..."
	@heighliner build --chain noble-fiattokenfactory-simd --local
	@echo "✅ Completed build!"

test: test-e2e test-unit

test-e2e:
	@echo "🤖 Running e2e tests..."
	@cd e2e && GOWORK=off go test -timeout 0 -race -v ./...
	@echo "✅ Completed e2e tests!"

test-unit:
	@echo "🤖 Running unit tests..."
	@go test -coverprofile=cover.out -race -count=1 ./x/...
	@echo "✅ Completed unit tests!"
	@grep -v -f .covignore cover.out > cover.filtered.out && rm cover.out
	@echo "\n📝 Detailed coverage report, excluding files in .covignore:"
	@go tool cover -func cover.filtered.out
	@go tool cover -html cover.filtered.out -o cover.html && rm cover.filtered.out
	@echo "\n📝 Produced html coverage report at cover.html, excluding files in .covignore"
