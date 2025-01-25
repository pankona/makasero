# MVP（Minimum Viable Product）設計書

## 1. 概要

### 1.1 目的
Roo Code Vimプラグインの最小機能セットとして、以下を実装します：
- インタラクティブなチャット機能
- ファイルコンテキストを利用した質問応答機能

### 1.2 機能範囲
- チャットインターフェース
- ファイル内容の読み取りと解析
- AIによる応答生成
- 会話の継続性の維持

## 2. ユーザーインターフェース

### 2.1 基本コマンド
```vim
" チャットの開始
:RooChat              " 通常のチャットを開始
:RooChatFile {file}   " ファイルについてのチャットを開始
:RooAddFile {file}    " 現在のチャットにファイルを追加
```

### 2.2 チャットバッファのレイアウト
```
+------------------------------------------+
|  Roo Chat                              x  |
|------------------------------------------+
| System: Code mode                         |
| Current context: example.py               |
|------------------------------------------+
| User: Please explain this file           |
|                                          |
| Assistant: This file contains a class    |
| that implements...                       |
|                                          |
| User: What is the purpose of the         |
| calculate method?                        |
|                                          |
| [Type your message below]                |
| ---------------------------------------- |
| >                                        |
+------------------------------------------+
```

### 2.3 キーマッピング
```vim
" デフォルトマッピング
nnoremap <leader>rc :RooChat<CR>        " チャット開始
nnoremap <leader>rf :RooChatFile %<CR>  " 現在のファイルについてチャット

" チャットバッファ内
i, a      " 入力開始
<CR>      " メッセージ送信
<C-c>     " 入力キャンセル
q         " チャットを閉じる
```

## 3. 技術実装

### 3.1 Vimプラグイン構造
```
autoload/roo/
├── chat.vim    " チャット機能のコア
├── buffer.vim  " バッファ管理
├── api.vim     " API通信
└── ui.vim      " UI表示
```

### 3.2 Goバックエンド構造
```
cmd/roo-helper/
├── main.go     " エントリーポイント
├── api/        " API通信
├── models/     " データモデル
└── vim/        " Vim連携
```

### 3.3 データモデル
```go
// チャットメッセージ
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ファイルコンテキスト
type FileContext struct {
    Path     string `json:"path"`
    Content  string `json:"content"`
    FileType string `json:"filetype"`
}

// チャットリクエスト
type ChatRequest struct {
    Message   string        `json:"message"`
    Context   []FileContext `json:"context"`
    History   []ChatMessage `json:"history"`
}
```

## 4. 実装フロー

### 4.1 フェーズ1: 基本構造（2-3日）
1. Goバックエンド
   - OpenAI API通信
   - 基本的なプロンプト処理
   - エラーハンドリング

2. Vimプラグイン基礎
   - チャットバッファ作成
   - 基本的なUI表示
   - キーマッピング

### 4.2 フェーズ2: コア機能（2-3日）
1. ファイルコンテキスト
   - ファイル読み込み
   - コンテキスト管理
   - プロンプト構築

2. チャット機能
   - メッセージ送受信
   - 履歴管理
   - 非同期処理

### 4.3 フェーズ3: UI/UX改善（1-2日）
1. 表示の改善
   - シンタックスハイライト
   - スクロール処理
   - ステータス表示

2. 操作性の向上
   - エラーメッセージの改善
   - コマンド補完
   - ヘルプドキュメント

## 5. テスト計画

### 5.1 ユニットテスト
```vim
" Vim側のテスト
function! Test_chat_buffer_creation() abort
    " バッファ作成のテスト
endfunction

function! Test_file_context() abort
    " ファイルコンテキストのテスト
endfunction
```

```go
// Go側のテスト
func TestChatRequest(t *testing.T) {
    // リクエスト処理のテスト
}

func TestPromptBuilder(t *testing.T) {
    // プロンプト構築のテスト
}
```

### 5.2 統合テスト
- Vim-Go間の通信テスト
- エンドツーエンドのシナリオテスト
- エラーケースのテスト

## 6. 成功基準

### 6.1 機能要件
- チャットバッファが正しく動作する
- ファイルの内容が正しく解析される
- AIの応答が適切に表示される
- 会話の文脈が維持される

### 6.2 非機能要件
- レスポンス時間が1秒以内
- メモリ使用が適切
- エラーが適切に処理される

### 6.3 品質基準
- テストカバレッジ80%以上
- エラーメッセージが明確
- ドキュメントが完備

## 7. 将来の拡張性

### 7.1 追加予定の機能
1. モードシステム
2. コマンド実行機能
3. ブラウザ操作機能
4. MCPサーバー連携

### 7.2 設計上の考慮点
- 機能の段階的な追加が容易
- 設定のカスタマイズ性
- プラグインAPIの提供