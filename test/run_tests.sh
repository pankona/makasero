#!/bin/bash

# エラー時に停止
set -e

# 現在のディレクトリを保存
CURRENT_DIR=$(pwd)

# テストディレクトリに移動
cd "$(dirname "$0")"

# Themisが存在するか確認
if [ ! -d "./themis" ]; then
    echo "Themisをインストールしています..."
    git clone https://github.com/thinca/vim-themis.git themis
fi

# プラグインのルートディレクトリを取得
PLUGIN_DIR="$(cd .. && pwd)"

# テスト用のvimrcを使用してテストを実行
echo "テストを実行中..."
THEMIS_VIM=vim
THEMIS_HOME=./themis
THEMIS_ARGS="-c \"source ./test.vimrc\""

./themis/bin/themis \
    --runtimepath "${PLUGIN_DIR}" \
    --runtimepath ./themis \
    --reporter spec \
    test_api.vim

# 元のディレクトリに戻る
cd "$CURRENT_DIR"