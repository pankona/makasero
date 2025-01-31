#!/bin/bash

# 色の定義
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# テスト用の一時ファイル
TEST_FILE="test/integration/test.go"

# テストカウンター
TESTS_TOTAL=0
TESTS_PASSED=0

# テスト関数
run_test() {
    local name=$1
    local cmd=$2
    local expected_status=$3
    
    echo -e "\n${BLUE}=== Running test: $name ===${NC}"
    echo "Command: $cmd"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # コマンドを実行
    echo -e "\n${BLUE}Response:${NC}"
    eval "$cmd" | tee /tmp/test_output
    local status=${PIPESTATUS[0]}
    
    # 期待する終了ステータスと比較
    echo -e "\n${BLUE}Result:${NC}"
    if [ $status -eq $expected_status ]; then
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC}"
        echo "Expected status: $expected_status, got: $status"
    fi
    echo -e "${BLUE}==============================${NC}\n"
}

# テスト用のファイルを作成
mkdir -p test/integration
cat > "$TEST_FILE" << 'EOF'
package main

func main() {
    // 改善が必要なコード
    x := 1
    if x == 1 {
        println("x is one")
    }
}
EOF

echo -e "${BLUE}Starting CLI integration tests...${NC}"

# ビルドテスト
echo -e "\n${BLUE}Building makasero...${NC}"
make build

# 各機能のテスト
run_test "Explain file" "./bin/makasero explain $TEST_FILE" 0

run_test "Explain code" "./bin/makasero explain 'func hello() { println(\"hello\") }'" 0

run_test "Chat simple" "./bin/makasero chat 'Goのスライスとは何ですか？'" 0

run_test "Chat JSON" "./bin/makasero chat '[{\"role\":\"system\",\"content\":\"あなたはGoの専門家です\"},{\"role\":\"user\",\"content\":\"Goのインターフェースについて簡単に説明してください\"}]'" 0

# テスト結果の表示
echo -e "\n${BLUE}=== Test Results ===${NC}"
echo "Total tests: $TESTS_TOTAL"
echo "Passed tests: $TESTS_PASSED"
echo "Failed tests: $((TESTS_TOTAL - TESTS_PASSED))"
echo -e "${BLUE}===================${NC}\n"

# テスト用ファイルの削除
rm -f "$TEST_FILE"

# 全てのテストが成功したかどうかを終了ステータスで返す
[ $TESTS_PASSED -eq $TESTS_TOTAL ] 