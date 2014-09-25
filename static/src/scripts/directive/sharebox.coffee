App.directive 'sharebox', (ace, storage) ->
	restrict: 'C',
	link: (scope, element) ->
		window.sharebox = scope
		ta = element.find "textarea"
		#watch url
		scope.$root.$watch 'shareURL', (url) ->
			return unless url
			element.fadeIn()
			ta.val(url).focus().select() 
		#timer
		t = 0
		#watch select
		ta.on "focus", ->
			clearTimeout t
		#watch deselect
		ta.on "blur", ->
			t = setTimeout ->
				scope.$root.shareURL = null
				element.fadeOut()
			, 2000
			return
		return