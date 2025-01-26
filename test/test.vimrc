" test.vimrc
" Roo Code テスト用の設定ファイル

" Vi互換モードをオフ
set nocompatible

" 日本語の設定
set encoding=utf-8
set fileencoding=utf-8
set fileencodings=utf-8,cp932,euc-jp
set ambiwidth=double

" 基本設定
set nobackup
set noswapfile
set noundofile
set nowritebackup
set history=10000

" プラグインのパスを追加
let s:test_dir = expand('<sfile>:p:h')
let s:plugin_dir = fnamemodify(s:test_dir, ':h')

" runtimepathの設定
let &runtimepath = s:plugin_dir . ',' . &runtimepath
let &runtimepath .= ',' . s:test_dir . '/themis'

" Themisの設定
let g:themis#suite_separator = ' > '

" テストモードを有効化
let g:roo_test_mode = 1

" エラー音を無効化
let g:roo_error_bell = 0

" 必要な機能を有効化
syntax enable
filetype plugin indent on

" テスト用の環境変数を設定
let $OPENAI_API_KEY = 'test_api_key'

" テスト用のグローバル変数
let g:roo_test = {
    \ 'mock_job_id': 1234,
    \ 'mock_output': [],
    \ 'mock_errors': [],
    \ 'mock_exit_status': 0
    \ }

" テスト用のヘルパー関数
function! TestSetup() abort
    " バッファをクリア
    silent! %bwipeout!
    
    " モックデータをリセット
    let g:roo_test.mock_output = []
    let g:roo_test.mock_errors = []
    let g:roo_test.mock_exit_status = 0
endfunction

" テスト実行前の初期化
augroup TestInit
    autocmd!
    autocmd VimEnter * call TestSetup()
augroup END