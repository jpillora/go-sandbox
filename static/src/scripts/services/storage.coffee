App.factory 'storage', ->

  # datums.type 'storage', {
  #   save: ->
  #   remove: ->
  # }

  wrap = (ns, fn) ->
    ->
      arguments[0] = [ns,arguments[0]].join('-')
      fn.apply null, arguments

  storage =
    create: (ns) ->
      s = {}
      s[k] = wrap ns, fn for k, fn of storage
      s
    get: (key) ->
      str = localStorage.getItem key
      if str and str.substr(0,4) is "J$ON"
        return JSON.parse str.substr 4
      return str
    set: (key, val) ->
      if typeof val is 'object'
        val = "J$ON#{JSON.stringify(val)}"
      localStorage.setItem key, val
    del: (key) ->
      localStorage.removeItem key

  window.storage = storage

