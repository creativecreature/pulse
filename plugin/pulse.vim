if exists('g:loaded_pulse_client')
	finish
endif

let g:loaded_pulse_client = 1
let g:pulse_session_id = substitute(system('uuidgen'), '\n', '', '')

function! s:RequirePulseClient(host) abort
	return jobstart(['pulse-client'], {'rpc': v:true})
endfunction

call remote#host#Register('pulse-client', 'x', function('s:RequirePulseClient'))

call remote#host#RegisterPlugin('pulse-client', '0', [
			\ {'type': 'function', 'name': 'OnFocusGained', 'sync': 1, 'opts': {}},
			\ {'type': 'function', 'name': 'OpenFile', 'sync': 1, 'opts': {}},
			\ {'type': 'function', 'name': 'SendHeartbeat', 'sync': 1, 'opts': {}},
			\ {'type': 'function', 'name': 'EndSession', 'sync': 1, 'opts': {}},
			\ ])


" We only want track time for one instance nvim instance. We
" need to let the server know which one we have focused.
autocmd FocusGained * :call call("OnFocusGained", [g:pulse_session_id, expand('%:p'), &filetype])

" Send FocusGained when we enter.
autocmd VimEnter * :call call("OnFocusGained", [g:pulse_session_id, expand('%:p'), &filetype])

" Let the server know the path of the buffer. Its not a problem to send
" temporary buffers. The server will figure it out and exit early.
autocmd BufEnter * :call call("OpenFile", [g:pulse_session_id, expand('%:p'), &filetype])

" We are sending a heartbeart each time we write a buffer.
" This lets the server know that our session is still active.
autocmd BufWrite * :call call("SendHeartbeat", [g:pulse_session_id, expand('%:p'), &filetype])

" When we exit VIM we inform the server that our coding session has ended.
autocmd VimLeave * :call call("EndSession", [g:pulse_session_id, expand('%:p'), &filetype])

" Timer variable to control heartbeat frequency for cursor movement.
let s:heartbeat_timer = -1
" Flag to indicate if the cursor has moved
let s:cursor_moved = 0

function! s:StartHeartbeatTimer() abort
  " Stop the existing timer if it's running
  if s:heartbeat_timer != -1
    call timer_stop(s:heartbeat_timer)
  endif
  " Start a new timer that repeats every 10 seconds (-1 is used for infinite repeats)
  let s:heartbeat_timer = timer_start(10000, function('s:HeartbeatTimerCallback'), {'repeat': -1})
endfunction

function! s:HeartbeatTimerCallback(timer_id) abort
  if s:cursor_moved
    " Send the heartbeat
    call call("SendHeartbeat", [g:pulse_session_id, expand('%:p'), &filetype])
    " Reset the cursor moved flag
    let s:cursor_moved = 0
  endif
endfunction

" Set the cursor moved flag when the cursor moves in normal mode
autocmd CursorMoved * let s:cursor_moved = 1

" Set the cursor moved flag when the cursor moves in insert mode
autocmd CursorMovedI * let s:cursor_moved = 1

autocmd VimEnter * call s:StartHeartbeatTimer()
