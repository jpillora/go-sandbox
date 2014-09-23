
App.controller 'Controls', ($rootScope, $scope, $window, ace, storage, key) ->

  scope = $rootScope.controls = $scope

  #bind run shortcut
  key.bind ['both+enter','shift+enter'], ->
    scope.run()

  scope.run = ->
    console.log "run!"


