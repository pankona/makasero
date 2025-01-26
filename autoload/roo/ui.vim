" autoload/roo/ui.vim
" Roo Code UIコンポーネントを管理する

let s:save_cpo = &cpo
set cpo&vim

" 内部変数
let s:buffer_name = '[Roo]'
let s:progress_timer = -1
let s:progress_frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']
let s:progress_index = 0
let s:progress_msg = ''

" バッファ管理関数
function! s:ensure_buffer() abort
    " 既存のバッファを探す
    let l:buf = bufnr(s:buffer_name)
    
    if l:buf == -1
        " バッファが存在しない場合は新規作成
        execute 'silent! new ' . s:buffer_name
        let l:buf = bufnr('%')
        
        " バッファ設定
        setlocal buftype=nofile
        setlocal bufhidden=hide
        setlocal noswapfile
        setlocal nobuflisted
        setlocal nowrap
        setlocal nospell
        setlocal nonumber
        setlocal norelativenumber
        setlocal nocursorline
        setlocal nomodifiable
        
        " シンタックスハイライト
        syntax match RooError /^Error:/
        syntax match RooProgress /^Progress:/
        highlight link RooError Error
        highlight link RooProgress Special
        
        " キーマッピング
        nnoremap <buffer> q :close<CR>
    else
        " 既存のバッファがある場合はそれを使用
        let l:win = bufwinnr(l:buf)
        if l:win == -1
            execute 'silent! split'
            execute 'buffer ' . l:buf
        else
            execute l:win . 'wincmd w'
        endif
    endif
    
    return l:buf
endfunction

" 出力表示
function! roo#ui#show_output(msg) abort
    let l:buf = s:ensure_buffer()
    
    " バッファを編集可能に
    setlocal modifiable
    
    " 内容をクリアして新しい出力を追加
    silent! %delete _
    call append(0, split(a:msg, '\n'))
    
    " 最初の行に移動して余分な行を削除
    normal! gg
    if line('$') > 1
        normal! Gdd
    endif
    
    " バッファを読み取り専用に
    setlocal nomodifiable
    
    " カーソルを先頭に
    normal! gg
endfunction

" プログレス表示
function! roo#ui#show_progress(msg) abort
    let s:progress_msg = a:msg
    let s:progress_index = 0
    
    " 既存のタイマーをクリア
    if s:progress_timer != -1
        call timer_stop(s:progress_timer)
    endif
    
    " プログレス表示を更新
    call s:update_progress()
    
    " タイマーを開始
    let s:progress_timer = timer_start(100, function('s:progress_callback'), {'repeat': -1})
endfunction

" プログレス表示を隠す
function! roo#ui#hide_progress() abort
    if s:progress_timer != -1
        call timer_stop(s:progress_timer)
        let s:progress_timer = -1
    endif
    
    let l:buf = bufnr(s:buffer_name)
    if l:buf != -1
        let l:win = bufwinnr(l:buf)
        if l:win != -1
            " 最後の出力を維持
            call s:ensure_buffer()
            setlocal modifiable
            call setline(1, '')
            setlocal nomodifiable
        endif
    endif
endfunction

" プログレス表示のコールバック
function! s:progress_callback(timer) abort
    call s:update_progress()
endfunction

" プログレス表示の更新
function! s:update_progress() abort
    let l:frame = s:progress_frames[s:progress_index]
    let s:progress_index = (s:progress_index + 1) % len(s:progress_frames)
    
    let l:buf = s:ensure_buffer()
    setlocal modifiable
    call setline(1, 'Progress: ' . l:frame . ' ' . s:progress_msg)
    setlocal nomodifiable
    
    " 画面を更新
    redraw
endfunction

" エラー表示
function! roo#ui#show_error(msg) abort
    let l:buf = s:ensure_buffer()
    
    " バッファを編集可能に
    setlocal modifiable
    
    " エラーメッセージを追加
    call append(0, 'Error: ' . a:msg)
    
    " 最初の行に移動して余分な行を削除
    normal! gg
    if line('$') > 1
        normal! Gdd
    endif
    
    " バッファを読み取り専用に
    setlocal nomodifiable
    
    " カーソルを先頭に
    normal! gg
    
    " エラー音を鳴らす（設定されている場合）
    if exists('g:roo_error_bell') && g:roo_error_bell
        execute "normal! \<Esc>"
    endif
endfunction

let &cpo = s:save_cpo
unlet s:save_cpo