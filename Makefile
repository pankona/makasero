.PHONY: build test clean install integration-test

# ビルド設定
BINARY_NAME=makasero
BUILD_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/makasero

test:
	@echo "Running tests..."
	@go test -v ./...

integration-test: build
	@echo "Running integration tests..."
	@./test/integration/test_cli.sh

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

install:
	@echo "Installing..."
	@go install ./cmd/makasero

# デフォルトターゲット
all: clean build

# 開発用のセットアップ
dev-setup:
	@go mod tidy
	@go mod verify