# アーキテクチャ設計

## 全体構造

プラグインは以下の主要コンポーネントで構成されます：

```
hello-vim-plugin-2/
├── plugin/
│   └── roo.vim          # プラグインのエントリーポイント
├── autoload/
│   └── roo/
│       ├── mode.vim     # モード管理
│       ├── api.vim      # AI API通信
│       ├── file.vim     # ファイル操作
│       ├── command.vim  # コマンド実行
│       ├── browser.vim  # ブラウザ操作
│       ├── mcp.vim      # MCPサーバー連携
│       ├── task.vim     # タスク管理
│       ├── config.vim   # 設定管理
│       └── ui.vim       # UI管理
└── doc/
    └── roo.txt          # ヘルプドキュメント
```

## コンポーネントの役割

### 1. コア機能

#### mode.vim
- モードの状態管理
- モード切り替え処理
- アクセス制御の実装
- カスタムモードの管理

```vim
" モードの定義
let s:modes = {
  \ 'code': {
    \ 'name': 'Code',
    \ 'groups': ['read', 'edit', 'browser', 'command', 'mcp']
  \ },
  \ 'architect': {
    \ 'name': 'Architect',
    \ 'groups': ['read', 'edit_md', 'browser', 'mcp']
  \ },
  \ 'ask': {
    \ 'name': 'Ask',
    \ 'groups': ['read', 'edit_md', 'browser', 'mcp']
  \ }
\ }

" 現在のモード管理
let s:current_mode = 'code'
```

#### api.vim
- AI APIとの通信処理
- プロンプト管理
- レスポンスのパース
- ストリーミング処理

```vim
" API設定
let s:api_config = {
  \ 'endpoint': '',
  \ 'model': '',
  \ 'api_key': ''
\ }

" APIリクエスト処理
function! roo#api#request(prompt) abort
  " 非同期リクエスト処理
endfunction
```

#### file.vim
- ファイル読み書き
- 差分適用
- ファイル検索
- パス解決

```vim
" ファイル操作関数
function! roo#file#read(path) abort
endfunction

function! roo#file#write(path, content) abort
endfunction

function! roo#file#apply_diff(path, diff) abort
endfunction
```

### 2. 拡張機能

#### command.vim
- コマンド実行の管理
- 出力のキャプチャ
- 非同期実行制御

```vim
" コマンド実行管理
function! roo#command#execute(cmd) abort
  " 非同期コマンド実行
endfunction

function! roo#command#capture_output() abort
  " 出力キャプチャ
endfunction
```

#### browser.vim
- ブラウザ制御
- スクリーンショット
- アクション実行

```vim
" ブラウザ操作
function! roo#browser#launch(url) abort
endfunction

function! roo#browser#execute_action(action, params) abort
endfunction
```

#### mcp.vim
- MCPサーバーとの通信
- ツール/リソース管理
- カスタムツール対応

```vim
" MCPサーバー連携
function! roo#mcp#connect(server) abort
endfunction

function! roo#mcp#execute_tool(tool, params) abort
endfunction
```

### 3. 補助機能

#### task.vim
- タスク状態管理
- 履歴保存/復元
- 中断/再開処理

```vim
" タスク管理
function! roo#task#start(task) abort
endfunction

function! roo#task#save_history() abort
endfunction

function! roo#task#resume(task_id) abort
endfunction
```

#### config.vim
- 設定ファイル管理
- デフォルト値設定
- 設定の検証

```vim
" 設定管理
function! roo#config#load() abort
endfunction

function! roo#config#validate() abort
endfunction
```

#### ui.vim
- 出力バッファ管理
- プログレス表示
- エラー表示

```vim
" UI管理
function! roo#ui#show_output(content) abort
endfunction

function! roo#ui#show_error(message) abort
endfunction
```

## データフロー

1. ユーザーインタラクション
```
User Input -> mode.vim -> api.vim -> UI更新
```

2. ファイル操作
```
file.vim -> mode.vim（権限チェック） -> 実行 -> UI更新
```

3. タスク実行
```
task.vim -> api.vim -> 各種ツール実行 -> 結果収集 -> UI更新
```

## エラー処理

- 各コンポーネントで適切なエラーハンドリング
- エラーメッセージの一元管理
- ユーザーフレンドリーなエラー表示
- リカバリー処理の実装

## 設定システム

```vim
" デフォルト設定
let g:roo_default_config = {
  \ 'api_key': '',
  \ 'model': 'gpt-4',
  \ 'default_mode': 'code',
  \ 'history_file': expand('~/.vim/roo_history.json'),
  \ 'log_level': 'info'
\ }
```

## 拡張性

- カスタムモードの追加
- 新しいツールの統合
- MCPサーバーの拡張
- カスタムUIの実装

## テスト戦略

1. ユニットテスト
- 各コンポーネントの独立したテスト
- モック/スタブの活用

2. 統合テスト
- コンポーネント間の連携テスト
- エンドツーエンドのシナリオテスト

3. 自動テスト
- CIでの自動テスト実行
- カバレッジ計測

## パフォーマンス考慮事項

- 非同期処理の活用
- キャッシュの適切な利用
- メモリ使用量の最適化
- 大きなファイルの効率的な処理