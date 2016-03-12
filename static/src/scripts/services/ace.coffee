App.factory 'ace', ($rootScope, storage) ->
	#prefixed store
	storage = storage.create 'ace'
	scope = $rootScope.ace = $rootScope.$new true
	Range = ace.require('ace/range').Range
	Selection = ace.require('ace/selection').Selection
	window.Selection = Selection
	editor = ace.edit "ace"
	session = editor.getSession()
	scope._ace = ace
	scope._editor = editor
	scope._session = session

	editor.commands.on "beforeExec", (e) ->
		console.log 'command', e
		if e.command.name is "del"
			editor.execCommand("duplicateSelection")
			e.preventDefault()
			return
		return

	scope.bindKey = (name, key, fn) ->
		editor.commands.addCommand
			name: name
			bindKey:
				win: key.replace "Super", "Ctrl"
				mac: key.replace "Super", "Command"
				sender: 'editor|cli'
			exec: fn
		return

	#no workers
	session.setUseWorker(false)

	#apply new settings
	scope.config = (c) ->
		editor.setTheme "ace/theme/#{c.theme}" if c.theme
		editor.setShowPrintMargin c.printMargin if 'printMargin' of c
		session.setMode "ace/mode/#{c.mode}" if c.mode
		session.setTabSize c.tabSize if 'tabSize' of c
		session.setUseSoftTabs c.softTabs if 'softTabs' of c

	scope.set = (val) ->
		c = editor.getCursorPosition()
		session.setValue val
		root.ace._session.getSelection().moveCursorTo c.row, c.column

	scope.readonly = (val) ->
		editor.setReadOnly !!val

	scope.get = ->
		session.getValue()

	unhight = 0
	markers = []
	scope.highlight = (loc) ->
		clearTimeout unhight
		r = new Range(loc.row, loc.col, loc.row, loc.col+1)
		m = session.addMarker r, "ace_warning", "text", true
		markers.push m

	scope.unhighlight = ->
		clearTimeout unhight
		while markers.length
			m = markers.pop()
			session.removeMarker m
		return

	#apply default config
	scope.config
		theme: "chrome"
		mode: "golang"
		tabSize: 4
		softTabs: false
		printMargin: false

	#set default code
	scope.set storage.get('current-code') or "package main\n\nfunc main() {\n\tprintln(42)\n}"

	editor.on 'change', ->
		#changing the code triggers markers to be removed
		clearTimeout unhight
		unhight = setTimeout scope.unhighlight, 1000
		storage.set 'current-code', scope.get()

	scope
