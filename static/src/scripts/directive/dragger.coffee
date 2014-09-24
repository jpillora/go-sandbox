App.directive 'dragger', (ace, storage) ->
	restrict: 'C',
	link: (scope, element) ->

		resize = (percent) ->
			top = percent + "%"
			bot = (100-percent) + "%"
			$("#ace").css("height", top)
			ace._editor.resize()
			$("#dragger").css("top", top)
			$("#out").css("top", top).css("height", bot)

		init = storage.get 'dragger-percent'
		resize init if init

		dragging = false
		element.on "mousedown", (e) ->
			dragging = true
			e.preventDefault()
			return

		$(window).on "mousemove", (e) ->
			return unless dragging
			percent = 100*(e.pageY/$(window).height())
			storage.set 'dragger-percent', percent
			resize percent
			return

		$(window).on "mouseup", ->
			return unless dragging
			dragging = false
			return
		return