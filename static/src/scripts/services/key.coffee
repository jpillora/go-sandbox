App.factory 'key', () ->
  key = Object.create Mousetrap
  #extend bind
  key.bind = (keys, fn) ->
    unless keys instanceof Array
      keys = [keys]
    newkeys = []
    for k in keys
      if /both/.test k
        newkeys.push k.replace "both", "ctrl"
        newkeys.push k.replace "both", "command"
      else
        newkeys.push k
    Mousetrap.bind newkeys, fn

  key