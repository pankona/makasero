# Vim実装計画

## 1. アーキテクチャの対応関係

### VSCode版とVim版の対応
```
VSCode                    Vim
------------------------------------------
extension.ts          -> plugin/roo.vim
ClineProvider        -> autoload/roo/core.vim
Webview (React)      -> ポップアップウィンドウ/バッファ
State Management     -> g:roo_state変数
Tool Execution       -> autoload/roo/tools/*.vim
```

## 2. コアコンポーネント

### 2.1 プラグインのエントリーポイント（plugin/roo.vim）
```vim
if exists('g:loaded_roo')
  finish
endif
let g:loaded_roo = 1

" コマンド定義
command! -nargs=1 RooChat call roo#chat#start(<q-args>)
command! -nargs=1 RooChatFile call roo#chat#start_with_file(<q-args>)
command! RooMode call roo#mode#show()
```

### 2.2 チャットインターフェース（autoload/roo/chat.vim）
```vim
function! roo#chat#start(initial_prompt) abort
  " チャットバッファの作成
  let bufnr = s:create_chat_buffer()
  
  " 初期プロンプトの設定
  call s:set_initial_prompt(bufnr, a:initial_prompt)
  
  " バッファの設定
  call s:setup_chat_buffer(bufnr)
  
  " キーマッピング
  call s:setup_chat_mappings(bufnr)
endfunction

function! s:create_chat_buffer() abort
  " 新しいバッファを作成
  let bufnr = nvim_create_buf(v:false, v:true)
  
  " バッファ名の設定
  call nvim_buf_set_name(bufnr, 'Roo Chat')
  
  " バッファオプションの設定
  call nvim_buf_set_option(bufnr, 'buftype', 'nofile')
  call nvim_buf_set_option(bufnr, 'swapfile', v:false)
  
  return bufnr
endfunction
```

### 2.3 状態管理（autoload/roo/state.vim）
```vim
" グローバル状態
let g:roo_state = {
  \ 'current_mode': 'code',
  \ 'chat_history': [],
  \ 'api_key': '',
  \ 'model': 'gpt-4'
  \ }

function! roo#state#get(key) abort
  return get(g:roo_state, a:key, v:null)
endfunction

function! roo#state#set(key, value) abort
  let g:roo_state[a:key] = a:value
endfunction
```

### 2.4 APIクライアント（autoload/roo/api.vim）
```vim
function! roo#api#send_message(message) abort
  " Goバックエンドへのリクエスト
  let job = job_start(['roo-helper', '-command', 'chat'],
    \ {'callback': function('s:handle_response')})
  
  " メッセージの送信
  call ch_sendraw(job, json_encode(a:message))
endfunction

function! s:handle_response(channel, msg) abort
  " レスポンスの処理
  let response = json_decode(a:msg)
  call s:update_chat_buffer(response)
endfunction
```

## 3. チャットUIの実装

### 3.1 バッファレイアウト
```
+------------------------------------------+
|  Roo Chat                              x  |
|------------------------------------------+
| System: Code mode                         |
| Current file: example.py                  |
|------------------------------------------+
| User: Please explain this file           |
|                                          |
| Assistant: This file contains a class    |
| that implements...                       |
|                                          |
| [Type your message below]                |
| ---------------------------------------- |
| >                                        |
+------------------------------------------+
```

### 3.2 キーマッピング
```vim
function! s:setup_chat_mappings(bufnr) abort
  " 入力モード
  call nvim_buf_set_keymap(a:bufnr, 'i', '<CR>', 
    \ '<CMD>call roo#chat#send_message()<CR>', {})
  
  " ノーマルモード
  call nvim_buf_set_keymap(a:bufnr, 'n', 'q',
    \ '<CMD>call roo#chat#close()<CR>', {})
  call nvim_buf_set_keymap(a:bufnr, 'n', '<C-r>',
    \ '<CMD>call roo#chat#apply_code()<CR>', {})
endfunction
```

## 4. ファイルコンテキスト機能

### 4.1 ファイル読み込み
```vim
function! roo#file#read_current() abort
  " 現在のバッファの内容を取得
  let content = getline(1, '$')
  let filetype = &filetype
  let filename = expand('%:p')
  
  return {
    \ 'content': content,
    \ 'filetype': filetype,
    \ 'filename': filename
    \ }
endfunction
```

### 4.2 コンテキスト管理
```vim
function! roo#context#add_file(file_info) abort
  " コンテキストの追加
  let context = {
    \ 'type': 'file',
    \ 'path': a:file_info.filename,
    \ 'content': a:file_info.content,
    \ 'filetype': a:file_info.filetype
    \ }
  
  call add(g:roo_state.context, context)
endfunction
```

## 5. 実装手順

### フェーズ1: 基本機能（2-3日）
1. プラグインの基本構造
   - エントリーポイント
   - 設定システム
   - バッファ管理

2. チャットインターフェース
   - バッファ作成
   - レイアウト
   - キーマッピング

3. Goバックエンド連携
   - API通信
   - レスポンス処理
   - エラーハンドリング

### フェーズ2: ファイルコンテキスト（2-3日）
1. ファイル操作
   - 現在のファイル読み込み
   - コンテキスト管理
   - 履歴管理

2. UI改善
   - シンタックスハイライト
   - スクロール処理
   - エラー表示

### フェーズ3: 機能拡張（2-3日）
1. 高度な機能
   - コード適用
   - ファイル保存
   - 設定カスタマイズ

2. ドキュメント
   - ヘルプファイル
   - READMEの更新
   - エラーメッセージの整備

## 6. テスト計画

### 6.1 ユニットテスト
```vim
function! Test_chat_buffer_creation() abort
  " バッファ作成のテスト
endfunction

function! Test_file_context() abort
  " ファイルコンテキストのテスト
endfunction
```

### 6.2 統合テスト
```vim
function! Test_chat_workflow() abort
  " チャットワークフローのテスト
endfunction

function! Test_api_communication() abort
  " API通信のテスト
endfunction
```

## 7. 成功基準

1. 基本機能
- チャットバッファが正しく動作する
- メッセージの送受信が機能する
- ファイルコンテキストが正しく追加される

2. ユーザビリティ
- 直感的な操作が可能
- レスポンスが適切
- エラーメッセージが分かりやすい

3. 拡張性
- 新機能の追加が容易
- カスタマイズが可能
- プラグインAPIの提供