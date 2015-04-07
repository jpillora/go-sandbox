App.factory 'render', (ace) ->
	contents = document.getElementById "contents"

	clear = ->
		contents.innerHTML = ''

	write = (type, msg) ->
		if /\u000c([^\u000c]*)$/.test msg
			msg = RegExp.$1
			clear()
		span = document.createElement "span"
		span.className = type
		$(span).text msg
		contents.appendChild span
		return

	handleErrors = (errstr) ->
		for err in errstr.split "\n"
			#empty error
			continue unless err
			#error on particular line
			if /\.go:(\d+):((\d+):)?\ (.+)$/.test err
				row = parseInt(RegExp.$1, 10)
				optcol = RegExp.$2
				optcol = " Col #{optcol}" if optcol
				col = RegExp.$3
				msg = RegExp.$4
				ace.highlight {row:row-1,col}
				write 'err', "Line #{row}:#{optcol} #{msg}\n"
			#write raw error
			else
				write 'err', "#{err}\n"
		return

	timer = 0
	handleEvents = (events, i = 0) ->
		clearTimeout timer
		if i is events.length
			write "exit", "\nProgram exited."
			return
		#peek
		e = events[i]
		if e.Delay
			ms = e.Delay/1000000
			e.Delay = 0
			timer = setTimeout ->
				handleEvents(events, i)
			, ms
			return
		type = if /^std(out|err)$/.test(e.Kind) then RegExp.$1 else 'out'
		write type, e.Message
		handleEvents(events, i+1)
		return

	render = (data) ->
		return unless data
		clear()
		clearTimeout timer
		handleErrors(data.Errors) if data.Errors
		handleEvents(data.Events) if data.Events

	return render
