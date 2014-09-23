App.factory 'console', ->

  ga 'create', 'UA-38709761-12', window.location.hostname
  ga 'send', 'pageview'

  str = (args) ->
    Array::slice.call(args).join(' ')

  log: ->
    console.log.apply console, arguments
    ga 'send', 'event', 'Log', str arguments

  error: ->
    console.error.apply console, arguments
    ga 'send', 'event', 'Error', str arguments
