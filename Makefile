.PHONY: build test clean install

# ビルド設定
BINARY_NAME=roo
BUILD_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/roo

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

install:
	@echo "Installing..."
	@go install ./cmd/roo

# デフォルトターゲット
all: clean build

# 開発用のセットアップ
dev-setup:
	@go mod tidy
	@go mod verify