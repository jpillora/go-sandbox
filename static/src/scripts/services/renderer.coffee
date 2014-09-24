App.factory 'render', () ->
	contents = document.getElementById "contents"

	clear = ->
		contents.innerHTML = ''


	write = (type, msg) ->
		if /\u000c([^\u000c]*)$/.test msg
			msg = RegExp.$1
			clear()

		# console.log "write [%s] %s", type, msg
		span = document.createElement "span"
		span.className = type
		span.innerText = msg
		contents.appendChild span
		return

	handleErrors = (errstr) ->
		errs = errstr.split "\n"
		errs.unshift()
		for err in errs
			#empty error
			continue unless err
			#error on particular line
			if /^prog\.go:(\d+):((\d+):)?\ (.+)$/.test err
				row = RegExp.$1
				col = RegExp.$3
				msg = RegExp.$4
				console.log "#%s %s => %s", row, col, msg
				write 'err', err
			#last line
			else if err is "[process exited with non-zero status]"
				write 'err', err
			else
				console.error "unknown error: %s", err
		return

	timer = 0
	handleEvents = (events) ->
		clearTimeout timer
		if events.length is 0
			write "exit", "\nProgram exited."
			return

		next = handleEvents.bind null, events
		#peek
		e = events[0]
		if e.Delay
			ms = e.Delay/1000000
			e.Delay = 0
			timer = setTimeout next, ms
			return

		write 'out', e.Message
		events.shift()
		next()
		return

	render = (data) ->
		return unless data
		clear()
		handleErrors(data.Errors) if data.Errors
		handleEvents(data.Events) if data.Events

	return render
