App.controller 'Controls', ($rootScope, $scope, $window, ace, storage, key, $http, render, console) ->

	scope = $rootScope.controls = $scope
	scope.super = if /Mac|iPod|iPhone|iPad/.test navigator.userAgent then "⌘" else "Ctrl"

	#bind run shortcut
	key.bind ['both+enter','shift+enter'], -> scope.compile()
	key.bind ['both+\\'], -> scope.imports()
	key.bind ['both+.'], -> scope.share()

	scope.compile = ->
		$rootScope.loading = true
		$http(
			method: 'POST',
			url: "/compile",
			data: $.param({version: 2, body: ace.get()}),
			headers: {'Content-Type': 'application/x-www-form-urlencoded'}
		).then (resp) ->
			console.log 'compiled'
			if resp.data.compile_errors
				render {Errors: resp.data.compile_errors}
			else
				render resp.data
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
			console.log 'formatted'
			ace.set(resp.data)
		.catch (resp) ->
			render { Errors: resp.data, Events: null } if resp.data
		.finally ->
			$rootScope.loading = false
			ace.readonly false

	loc = window.location
	$rootScope.shareURL = null
	scope.share = ->
		$rootScope.loading = true
		$http.post("/share", ace.get()).then (resp) ->
			console.log 'shared'
			$rootScope.shareURL = loc.protocol + "//" + loc.host + "/#/" + resp.data
		.catch (resp) ->
			console.error "share failed, oh noes", resp
		.finally ->
			$rootScope.loading = false

	loadShare = (id) ->
		ace.readonly true
		$rootScope.loading = true
		$http.get("/p/#{id}").then (resp) ->
			str = resp.data
			ta = $ str.substring(str.indexOf("<textarea"), str.indexOf("</textarea>")+12)
			ace.set(ta.val())
		.catch (resp) ->
			console.error "load share failed, oh noes", resp
		.finally ->
			ace.readonly false
			$rootScope.loading = false

	hash = null
	checkHash = ->
		return if hash is loc.hash
		hash = loc.hash
		return unless /#\/([\w-]+)$/.test hash
		loadShare RegExp.$1

	setInterval checkHash, 1000