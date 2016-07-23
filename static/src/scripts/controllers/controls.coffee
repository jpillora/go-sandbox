App.controller 'Controls', ($rootScope, $scope, $window, ace, storage, $http, render, console) ->

	scope = $rootScope.controls = $scope
	scope.super = if /Mac|iPod|iPhone|iPad/.test navigator.userAgent then "⌘" else "Ctrl"

	#bind run shortcut
	ace.bindKey "compile", "Super-Enter", -> scope.importsCompile()
	ace.bindKey "imports", "Super-\\", -> scope.imports()
	ace.bindKey "share", "Super-.", -> scope.share()
	ace.bindKey "save", "Super-s", -> scope.save()
	ace.bindKey "duplicate", "Ctrl-d", -> ace._editor.execCommand("duplicateSelection")

	scope.importsCompile = ->
		return if $rootScope.loading
		$rootScope.loading = true
		ace.readonly true
		$http.post("/importscompile", ace.get()).then (resp) ->
			console.log 'compiled'
			# if resp.data.compile_errors
			# 	render {Errors: resp.data.compile_errors, Events: null}
			# else
			render resp.data
			if resp.data.NewCode
				ace.set(resp.data.NewCode)
			return
		.catch (err) ->
			console.error "imports/compile failed, oh noes", err
		.finally ->
			$rootScope.loading = false
			ace.readonly false

	scope.imports = ->
		return if $rootScope.loading
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
		return if $rootScope.loading
		$rootScope.loading = true
		$http.post("/share", ace.get()).then (resp) ->
			console.log 'shared'
			$rootScope.shareURL = loc.protocol + "//" + loc.host + "/#/" + resp.data
		.catch (resp) ->
			console.error "share failed, oh noes", resp
		.finally ->
			$rootScope.loading = false

	loadShare = (id) ->
		return if $rootScope.loading
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
	checkHash()
	setInterval checkHash, 1000

	scope.save = ->
		return unless scope.save.supported
		a = document.createElementNS("http://www.w3.org/1999/xhtml", "a")
		blob = new window.Blob([ace.get()], type: "text/plain;charset=utf8")
		a.href = window.URL.createObjectURL(blob)
		a.download = "snippet-"+(++scope.save.s)+".go"
		event = document.createEvent "MouseEvents"
		event.initMouseEvent "click", 1, 0, window, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, null
		a.dispatchEvent event
		return
	scope.save.s = 0 #saves
	scope.save.supported = "download" of document.createElementNS("http://www.w3.org/1999/xhtml", "a")
