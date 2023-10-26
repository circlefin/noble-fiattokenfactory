.PHONY: proto-setup proto-format proto-lint proto-gen format lint test-e2e test-unit build
all: proto-all format lint test-unit build

###############################################################################
###                                  Build                                  ###
###############################################################################

build:
	@echo "🤖 Building simd..."
	@cd simapp && make build
	@echo "✅ Completed build!"

###############################################################################
###                          Formatting & Linting                           ###
###############################################################################

gofumpt_cmd=mvdan.cc/gofumpt
golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint

format:
	@echo "🤖 Running formatter..."
	@go run $(gofumpt_cmd) -l -w .
	@echo "✅ Completed formatting!"

lint:
	@echo "🤖 Running linter..."
	@go run $(golangci_lint_cmd) run --timeout=10m
	@echo "✅ Completed linting!"

###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.27.1

proto-all: proto-format proto-lint proto-gen

proto-format:
	@echo "🤖 Running protobuf formatter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "✅ Completed protobuf formatting!"

proto-gen:
	@echo "🤖 Generating code from protobuf..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		noble-fiattokenfactory-proto sh ./proto/generate.sh
	@echo "✅ Completed code generation!"

proto-lint:
	@echo "🤖 Running protobuf linter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "✅ Completed protobuf linting!"

proto-setup:
	@echo "🤖 Setting up protobuf environment..."
	@docker build --rm --tag noble-fiattokenfactory-proto:latest --file proto/Dockerfile .
	@echo "✅ Setup protobuf environment!"

###############################################################################
###                                 Testing                                 ###
###############################################################################

heighliner:
	@echo "🤖 Building image..."
	@heighliner build --chain noble-fiattokenfactory-simd --local 1> /dev/null
	@echo "✅ Completed build!"

test: test-e2e test-unit

test-e2e:
	@echo "🤖 Running e2e tests..."
	@cd e2e && GOWORK=off go test -race -v ./...
	@echo "✅ Completed e2e tests!"

test-unit:
	@echo "🤖 Running unit tests..."
	@go test -cover -race -v ./x/...
	@echo "✅ Completed unit tests!"
