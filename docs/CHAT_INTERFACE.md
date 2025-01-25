# チャットインターフェース設計

## 1. UI設計

### 1.1 チャットバッファのレイアウト
```
+------------------------------------------+
|  Roo Chat                              x  |
|------------------------------------------+
| System: Code mode                         |
|------------------------------------------+
| User: Create a function that calculates   |
| fibonacci numbers                         |
|                                          |
| Assistant: Here's a function that         |
| calculates Fibonacci numbers:             |
|                                          |
| ```python                                |
| def fibonacci(n):                        |
|     if n <= 1:                           |
|         return n                         |
|     return fibonacci(n-1) + fibonacci(n-2)|
| ```                                      |
|                                          |
| This is a recursive implementation...     |
|                                          |
| [Type your message below]                |
| ---------------------------------------- |
| >                                        |
+------------------------------------------+
```

### 1.2 バッファ構成
- チャット履歴領域（読み取り専用）
- 入力領域（編集可能）
- ステータス領域（モード表示等）

## 2. 操作方法

### 2.1 チャットの開始
```vim
" コマンド
:RooChat                  " 新しいチャットを開始
:RooChat {initial_prompt} " プロンプトを指定してチャット開始
:RooChatBuffer           " 現在のバッファについてチャット開始

" キーマッピング
nnoremap <leader>rc :RooChat<CR>
nnoremap <leader>rb :RooChatBuffer<CR>
```

### 2.2 チャット操作
```vim
" 入力モード
i, a      " 入力開始
<CR>      " メッセージ送信（入力モード中）
<C-c>     " 入力キャンセル

" ノーマルモード
j, k      " チャット履歴のスクロール
gg        " チャット履歴の先頭へ
G         " チャット履歴の末尾へ
q         " チャットを閉じる
<C-r>     " コード提案を現在のバッファに適用
?         " ヘルプを表示
```

## 3. 実装詳細

### 3.1 バッファ管理
```vim
" autoload/roo/chat.vim
function! roo#chat#create() abort
    " チャットバッファの作成
    let bufnr = nvim_create_buf(v:false, v:true)
    call nvim_buf_set_name(bufnr, 'Roo Chat')
    
    " バッファ設定
    call nvim_buf_set_option(bufnr, 'buftype', 'nofile')
    call nvim_buf_set_option(bufnr, 'swapfile', v:false)
    
    " 領域の初期化
    call s:initialize_chat_regions(bufnr)
    
    return bufnr
endfunction

function! s:initialize_chat_regions(bufnr) abort
    " チャット履歴領域の設定
    call nvim_buf_set_lines(a:bufnr, 0, -1, v:true, [
        \ 'System: ' . roo#mode#get_current(),
        \ repeat('-', 42),
        \ '',
        \ '[Type your message below]',
        \ repeat('-', 40),
        \ '> '
    \ ])
    
    " 入力領域の設定
    call nvim_buf_set_option(a:bufnr, 'modifiable', v:true)
endfunction
```

### 3.2 メッセージ処理
```vim
function! roo#chat#send_message() abort
    " 入力内容の取得
    let input = s:get_input()
    if empty(input)
        return
    endif
    
    " メッセージの追加
    call s:append_message('User', input)
    
    " API呼び出し
    let response = roo#api#chat(input)
    call s:append_message('Assistant', response)
    
    " 入力領域のクリア
    call s:clear_input()
endfunction

function! s:append_message(role, content) abort
    let lines = split(a:content, '\n')
    let formatted = [a:role . ': ' . lines[0]]
    call extend(formatted, map(lines[1:], {_, v -> '    ' . v}))
    
    " 履歴領域に追加
    call s:insert_before_input(formatted)
endfunction
```

### 3.3 シンタックスハイライト
```vim
" syntax/roochat.vim
syntax match RooChatUser /^User: .*/
syntax match RooChatAssistant /^Assistant: .*/
syntax match RooChatSystem /^System: .*/
syntax match RooChatPrompt /^> .*/
syntax region RooChatCode start=/```\w*$/ end=/```/ contains=@NoSpell

highlight default link RooChatUser Statement
highlight default link RooChatAssistant Identifier
highlight default link RooChatSystem Comment
highlight default link RooChatPrompt Question
highlight default link RooChatCode String
```

### 3.4 コマンド定義
```vim
" plugin/roo.vim
command! -nargs=? RooChat call roo#chat#start(<q-args>)
command! RooChatBuffer call roo#chat#start_with_buffer()
```

## 4. Goバックエンドとの連携

### 4.1 チャット履歴の管理
```go
// cmd/roo-helper/models/chat.go
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatHistory struct {
    Messages []ChatMessage `json:"messages"`
    Mode    string        `json:"mode"`
}
```

### 4.2 APIリクエスト
```go
// cmd/roo-helper/api/chat.go
func HandleChatRequest(history ChatHistory, newMessage string) (string, error) {
    messages := append(history.Messages, ChatMessage{
        Role:    "user",
        Content: newMessage,
    })
    
    // APIリクエストの構築と送信
    response, err := sendChatRequest(messages)
    if err != nil {
        return "", err
    }
    
    return response, nil
}
```

## 5. 特殊機能

### 5.1 コードスニペットの適用
```vim
function! roo#chat#apply_code() abort
    " 最後のコードブロックを検索
    let code = s:find_last_code_block()
    if empty(code)
        return
    endif
    
    " 現在のバッファに適用
    let current_buf = bufnr('%')
    call nvim_buf_set_lines(current_buf, 0, -1, v:true, code)
endfunction
```

### 5.2 コンテキストメニュー
```vim
let g:roo_chat_context_menu = [
    \ ['Send Message', 'call roo#chat#send_message()'],
    \ ['Apply Code', 'call roo#chat#apply_code()'],
    \ ['Clear Chat', 'call roo#chat#clear()'],
    \ ['Change Mode', 'call roo#chat#change_mode()'],
    \ ]
```

### 5.3 履歴の保存と復元
```vim
function! roo#chat#save_history() abort
    let history = {
        \ 'messages': s:get_chat_messages(),
        \ 'mode': roo#mode#get_current(),
        \ 'timestamp': strftime('%Y-%m-%d %H:%M:%S')
        \ }
    
    " JSONとして保存
    call writefile([json_encode(history)], 
        \ expand('~/.vim/roo_chat_history.json'), 'a')
endfunction
```

## 6. エラーハンドリング

### 6.1 入力検証
```vim
function! s:validate_input(input) abort
    if empty(a:input)
        call roo#ui#warn('Message cannot be empty')
        return v:false
    endif
    return v:true
endfunction
```

### 6.2 API エラー処理
```vim
function! s:handle_api_error(error) abort
    let message = '[Roo] API Error: ' . a:error
    call s:append_message('System', message)
    call roo#ui#error(message)
endfunction
```

## 7. パフォーマンス最適化

### 7.1 バッファ更新の最適化
```vim
function! s:update_chat_buffer() abort
    " 更新が必要な領域のみを更新
    let start_line = s:find_update_start()
    let end_line = s:find_update_end()
    
    call nvim_buf_set_lines(s:chat_bufnr,
        \ start_line, end_line, v:true, s:new_content)
endfunction
```

### 7.2 非同期処理
```vim
function! roo#chat#send_message_async() abort
    " 非同期でAPIリクエストを送信
    call roo#job#start(s:get_input(), function('s:handle_response'))
endfunction
```

## 8. テスト計画

### 8.1 ユニットテスト
```vim
function! Test_chat_buffer_creation() abort
    " バッファ作成のテスト
endfunction

function! Test_message_formatting() abort
    " メッセージフォーマットのテスト
endfunction
```

### 8.2 統合テスト
```vim
function! Test_chat_workflow() abort
    " チャットワークフローの統合テスト
endfunction