App.factory 'console', ->

  ga('create', 'UA-38709761-13', 'auto')
  ga('send', 'pageview')

  setInterval (-> ga 'send', 'event', 'Ping'), 60*1000

  str = (args) ->
    Array::slice.call(args).join(' ')

  log: ->
    console.log.apply console, arguments
    ga 'send', 'event', 'Log', str arguments

  error: ->
    console.error.apply console, arguments
    ga 'send', 'event', 'Error', str arguments
