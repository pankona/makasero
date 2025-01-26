" autoload/roo/api.vim
" Roo Code APIとの連携を管理する

let s:save_cpo = &cpo
set cpo&vim

" 内部変数
let s:jobs = {}
let s:callbacks = {}

" APIリクエストを送信する
function! roo#api#send_request(command, input, callback) abort
    " リクエストデータの作成
    let l:request = {
        \ 'command': a:command,
        \ 'input': a:input,
        \ 'options': {}
        \ }

    " JSONエンコード
    let l:json_data = json_encode(l:request)

    " Go実行ファイルのパスを取得
    let l:helper_path = expand('$HOME/.vim/bin/roo-helper')
    if !executable(l:helper_path)
        call roo#ui#show_error('roo-helper executable not found at: ' . l:helper_path)
        return -1
    endif

    " コマンドの構築
    let l:cmd = [
        \ l:helper_path,
        \ '-command', a:command,
        \ '-input', a:input
        \ ]

    " ジョブオプションの設定
    let l:job_opts = {
        \ 'out_cb': function('s:handle_output'),
        \ 'err_cb': function('s:handle_error'),
        \ 'exit_cb': function('s:handle_exit'),
        \ 'mode': 'raw'
        \ }

    " 非同期ジョブの開始
    let l:job = job_start(l:cmd, l:job_opts)
    let l:job_id = job_info(l:job)['process']

    " ジョブとコールバックの保存
    let s:jobs[l:job_id] = l:job
    let s:callbacks[l:job_id] = a:callback

    return l:job_id
endfunction

" コード説明リクエストを送信
function! roo#api#explain_code(code) abort
    " プログレス表示の開始
    call roo#ui#show_progress('Analyzing code...')

    " リクエストの送信
    return roo#api#send_request('explain', a:code, function('s:handle_explanation'))
endfunction

" チャットリクエストを送信
function! roo#api#chat(messages) abort
    " プログレス表示の開始
    call roo#ui#show_progress('Processing chat...')

    " リクエストの送信
    return roo#api#send_request('chat', json_encode(a:messages), function('s:handle_chat'))
endfunction

" 出力ハンドラ
function! s:handle_output(channel, msg) abort
    let l:job = ch_getjob(a:channel)
    let l:job_id = job_info(l:job)['process']

    " 結果の解析
    try
        let l:response = json_decode(a:msg)
        if has_key(s:callbacks, l:job_id)
            call s:callbacks[l:job_id](l:response)
        endif
    catch
        call roo#ui#show_error('Failed to parse response: ' . v:exception)
    endtry
endfunction

" エラーハンドラ
function! s:handle_error(channel, msg) abort
    call roo#ui#show_error('Error: ' . a:msg)
endfunction

" 終了ハンドラ
function! s:handle_exit(channel, status) abort
    let l:job = ch_getjob(a:channel)
    let l:job_id = job_info(l:job)['process']

    " プログレス表示の終了
    call roo#ui#hide_progress()

    " ジョブの後処理
    if has_key(s:jobs, l:job_id)
        unlet s:jobs[l:job_id]
    endif
    if has_key(s:callbacks, l:job_id)
        unlet s:callbacks[l:job_id]
    endif

    " エラー状態の確認
    if a:status != 0
        call roo#ui#show_error('Process exited with status: ' . a:status)
    endif
endfunction

" コード説明の結果ハンドラ
function! s:handle_explanation(response) abort
    if !a:response.success
        call roo#ui#show_error('Failed to explain code: ' . a:response.error)
        return
    endif

    " 結果の表示
    call roo#ui#show_output(a:response.data)
endfunction

" チャットの結果ハンドラ
function! s:handle_chat(response) abort
    if !a:response.success
        call roo#ui#show_error('Chat error: ' . a:response.error)
        return
    endif

    " 結果の表示
    call roo#ui#show_output(a:response.data)
endfunction

let &cpo = s:save_cpo
unlet s:save_cpo