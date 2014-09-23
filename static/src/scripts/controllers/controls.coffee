
App.controller 'Controls', ($rootScope, $scope, $window, ace, storage, key, $http) ->

	scope = $rootScope.controls = $scope

	#bind run shortcut
	key.bind ['both+enter','shift+enter'], -> scope.run()
	key.bind ['both+s'], -> scope.imports()

	queueMessage = (msg) ->
		setTimeout ->
			$rootScope.$apply -> $rootScope.output += msg.Message
		, msg.Delay

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
			#last line
			else if err is "[process exited with non-zero status]"
				console.log " ==> %s", err
			else
				console.error "unknown error: %s", err
		return

	handleEvents = (events) ->
		for e in events
			if typeof e.Delay is "number" and e.Message
				queueMessage e
			else
				console.log "msg???", e
		return

	scope.run = ->
		$rootScope.loading = true
		$http.post("/compile", ace.get()).then (resp) ->
			$rootScope.output = ""
			handleErrors(resp.data.Errors) if resp.data.Errors
			handleEvents(resp.data.Events) if resp.data.Events
			return
		.catch (err) ->
			console.error "compile failed, oh noes", err
		.finally ->
			$rootScope.loading = false
			ace.readonly false

	scope.imports = ->
		$rootScope.loading = true
		ace.readonly true
		$http.post("/imports", ace.get()).then (resp) ->
			ace.set(resp.data)
		.catch (resp) ->
			handleErrors resp.data if resp.data
		.finally ->
			$rootScope.loading = false
			ace.readonly false

