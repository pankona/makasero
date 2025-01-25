.PHONY: build test clean

# 変数定義
BINARY_NAME=roo-helper
BUILD_DIR=../../bin

# ビルドターゲット
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)

# テストターゲット
test:
	@echo "Running tests..."
	@go test ./... -v

# クリーンターゲット
clean:
	@echo "Cleaning..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)

# 依存関係の更新
deps:
	@echo "Updating dependencies..."
	@go mod tidy

# インストールターゲット
install: build
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p ~/.vim/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.vim/bin/

# すべてのターゲットを実行
all: clean deps test build

# デフォルトターゲット
default: build