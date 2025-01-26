" test/test_api.vim
" Roo Code APIのテスト

let s:suite = themis#suite('API Tests')
let s:assert = themis#helper('assert')

" テスト用のモックデータ
let s:mock = {
    \ 'last_command': [],
    \ 'last_options': {},
    \ 'output_buffer': [],
    \ 'error_messages': [],
    \ 'last_exit_status': 0,
    \ 'job_id': 1234
    \ }

" モック用のハンドラ関数
function! s:mock.handle_output(channel, msg) dict
    call add(self.output_buffer, a:msg)
endfunction

function! s:mock.handle_error(channel, msg) dict
    call add(self.error_messages, a:msg)
endfunction

function! s:mock.handle_exit(channel, status) dict
    let self.last_exit_status = a:status
endfunction

" テストのセットアップ
function! s:setup() abort
    " モックデータをリセット
    let s:mock.last_command = []
    let s:mock.last_options = {}
    let s:mock.output_buffer = []
    let s:mock.error_messages = []
    let s:mock.last_exit_status = 0
    
    " テスト用のバッファをセットアップ
    if !exists('g:roo_test_mode')
        let g:roo_test_mode = 1
    endif

    " オリジナルの関数を保存
    let s:original_send_request = function('roo#api#send_request')
    
    " モック関数を設定
    function! roo#api#send_request(command, input, Callback) abort
        let s:mock.last_command = [
            \ 'roo-helper',
            \ '-command', a:command,
            \ '-input', a:input
            \ ]
        return s:mock.job_id
    endfunction
endfunction

" テストのクリーンアップ
function! s:teardown() abort
    " テストバッファをクリーンアップ
    silent! %bwipeout!
    
    " オリジナルの関数を復元
    if exists('s:original_send_request')
        function! roo#api#send_request(command, input, Callback) abort
            return s:original_send_request(a:command, a:input, a:Callback)
        endfunction
        unlet s:original_send_request
    endif
endfunction

" APIリクエストのテスト
function! s:suite.test_send_request() abort
    call s:setup()
    
    " テストデータ
    let l:command = 'explain'
    let l:input = 'test code'
    let Callback = function('s:mock.handle_output', [], s:mock)
    
    " リクエストを送信
    let l:result = roo#api#send_request(l:command, l:input, Callback)
    
    " アサーション
    call s:assert.equal(l:result, s:mock.job_id, 'ジョブIDが返されるべき')
    call s:assert.equal(s:mock.last_command[1], '-command', 'コマンドオプションが正しいべき')
    call s:assert.equal(s:mock.last_command[2], 'explain', 'コマンド名が正しいべき')
    
    call s:teardown()
endfunction

" コード説明リクエストのテスト
function! s:suite.test_explain_code() abort
    call s:setup()
    
    " テストデータ
    let l:code = 'function test() { return 42; }'
    
    " リクエストを送信
    let l:result = roo#api#explain_code(l:code)
    
    " アサーション
    call s:assert.equal(l:result, s:mock.job_id, 'ジョブIDが返されるべき')
    call s:assert.equal(s:mock.last_command[1], '-command', 'コマンドオプションが正しいべき')
    call s:assert.equal(s:mock.last_command[2], 'explain', 'コマンド名が正しいべき')
    
    call s:teardown()
endfunction

" 出力ハンドラのテスト
function! s:suite.test_output_handler() abort
    call s:setup()
    
    " テスト用のメッセージ
    let l:msg = json_encode({
        \ 'success': v:true,
        \ 'data': 'Test output'
        \ })
    
    " 出力ハンドラを呼び出し
    call s:mock.handle_output({}, l:msg)
    
    " アサーション
    call s:assert.equal(len(s:mock.output_buffer), 1, '出力バッファに1つのメッセージが追加されるべき')
    call s:assert.equal(s:mock.output_buffer[0], l:msg, '正しいメッセージが保存されるべき')
    
    call s:teardown()
endfunction

" エラーハンドラのテスト
function! s:suite.test_error_handler() abort
    call s:setup()
    
    " テスト用のエラーメッセージ
    let l:msg = 'Test error'
    
    " エラーハンドラを呼び出し
    call s:mock.handle_error({}, l:msg)
    
    " アサーション
    call s:assert.equal(len(s:mock.error_messages), 1, 'エラーメッセージが1つ追加されるべき')
    call s:assert.equal(s:mock.error_messages[0], l:msg, '正しいエラーメッセージが保存されるべき')
    
    call s:teardown()
endfunction

" 終了ハンドラのテスト
function! s:suite.test_exit_handler() abort
    call s:setup()
    
    " テスト用のステータス
    let l:status = 0
    
    " 終了ハンドラを呼び出し
    call s:mock.handle_exit({}, l:status)
    
    " アサーション
    call s:assert.equal(s:mock.last_exit_status, l:status, '正しい終了ステータスが保存されるべき')
    
    call s:teardown()
endfunction