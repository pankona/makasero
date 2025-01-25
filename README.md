# hello-vim-plugin-2

Roo Code VSCode拡張の機能をVimプラグインとして実装するプロジェクト。

## ドキュメント構成

### コア設計
- `docs/MVP.md` - 最小実装の機能定義
  - チャット機能
  - ファイルコンテキスト機能の仕様

- `docs/ARCHITECTURE.md` - 全体アーキテクチャ設計
  - システム構成
  - データフロー
  - コンポーネント間の関係

### 実装設計
- `docs/IMPLEMENTATION_PLAN.md` - 実装の詳細計画
  - Vimプラグインの構造
  - コンポーネント設計
  - 実装手順と優先順位

- `docs/GO_IMPLEMENTATION.md` - バックエンド設計
  - API通信の実装
  - プロンプト処理
  - エラーハンドリング

- `docs/CHAT_INTERFACE.md` - チャットUI設計
  - バッファレイアウト
  - キーマッピング
  - インタラクション設計

### ユーザードキュメント
- `doc/roo.txt` - Vimヘルプドキュメント
  - インストール方法
  - 使用方法
  - コマンドリファレンス

## 開発環境

- Vim 8.1以上
- Go 1.21以上
- curl（API通信用）

## インストール

```bash
# プラグインのインストール
git clone https://github.com/yourusername/hello-vim-plugin-2.git ~/.vim/pack/plugins/start/hello-vim-plugin-2

# Goバックエンドのビルド
cd ~/.vim/pack/plugins/start/hello-vim-plugin-2
go build -o bin/roo-helper cmd/roo-helper/main.go
```

## 使用方法

```vim
" チャットの開始
:RooChat

" ファイルについてのチャット
:RooChatFile %

" モードの切り替え
:RooMode
```

## ライセンス

MIT