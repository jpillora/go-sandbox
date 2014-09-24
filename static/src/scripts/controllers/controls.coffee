
App.controller 'Controls', ($rootScope, $scope, $window, ace, storage, key, $http, render) ->

	scope = $rootScope.controls = $scope
	scope.super = if /Mac|iPod|iPhone|iPad/.test navigator.userAgent then "⌘" else "Ctrl"

	#bind run shortcut
	key.bind ['both+enter','shift+enter'], -> scope.run()
	key.bind ['both+\\'], -> scope.imports()

	scope.run = ->
		$rootScope.loading = true
		$http.post("/compile", ace.get()).then (resp) ->
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
			ace.set(resp.data)
		.catch (resp) ->
			render { Errors: resp.data, Events: null } if resp.data
		.finally ->
			$rootScope.loading = false
			ace.readonly false

