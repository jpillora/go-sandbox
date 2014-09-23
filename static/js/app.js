(function() {
  var App;

  App = window.App = angular.module('sandbox', []);

  App.controller('Controls', function($rootScope, $scope, $window, ace, storage, key) {
    var scope;
    scope = $rootScope.controls = $scope;
    key.bind(['both+enter', 'shift+enter'], function() {
      return scope.run();
    });
    return scope.run = function() {
      return console.log("run!");
    };
  });

  App.controller('Output', function($rootScope) {
    return $rootScope.output = '';
  });

  App.factory('ace', function($rootScope, storage, key) {
    var Range, editor, scope, session;
    storage = storage.create('ace');
    scope = $rootScope.ace = $rootScope.$new(true);
    Range = ace.require('ace/range').Range;
    editor = ace.edit("ace");
    session = editor.getSession();
    scope._ace = ace;
    scope._editor = editor;
    scope._session = session;
    editor.setKeyboardHandler({
      handleKeyboard: function(data, hashId, keyString, keyCode, e) {
        var keys, str;
        if (!e) {
          return;
        }
        keys = [];
        if (e.ctrlKey) {
          keys.push('ctrl');
        }
        if (e.metaKey) {
          keys.push('command');
        }
        if (e.shiftKey) {
          keys.push('shift');
        }
        keys.push(keyString);
        str = keys.join('+');
        if (key.isBound(str)) {
          key.trigger(str);
          e.preventDefault();
          return false;
        }
        return true;
      }
    });
    session.setUseWorker(false);
    scope.config = function(c) {
      if (c.theme) {
        editor.setTheme("ace/theme/" + c.theme);
      }
      if ('printMargin' in c) {
        editor.setShowPrintMargin(c.printMargin);
      }
      if (c.mode) {
        session.setMode("ace/mode/" + c.mode);
      }
      if ('tabSize' in c) {
        session.setTabSize(c.tabSize);
      }
      if ('softTabs' in c) {
        return session.setUseSoftTabs(c.softTabs);
      }
    };
    scope.set = function(val) {
      return session.setValue(val);
    };
    scope.get = function() {
      return session.getValue();
    };
    scope.highlight = function(loc, t) {
      var m, r;
      if (t == null) {
        t = 3000;
      }
      r = new Range(loc.row, loc.col, loc.row, loc.col + 1);
      m = session.addMarker(r, "ace_warning", "text", true);
      return setTimeout(function() {
        return session.removeMarker(m);
      }, t);
    };
    scope.config({
      theme: "github",
      mode: "golang",
      tabSize: 2,
      softTabs: false,
      printMargin: false
    });
    scope.set(storage.get('current-code') || "package main\n\nfunc main() {\n\tprintln(42)\n}");
    editor.on('change', function() {
      return storage.set('current-code', scope.get());
    });
    return scope;
  });

  App.factory('console', function() {
    var str;
    ga('create', 'UA-38709761-12', window.location.hostname);
    ga('send', 'pageview');
    str = function(args) {
      return Array.prototype.slice.call(args).join(' ');
    };
    return {
      log: function() {
        console.log.apply(console, arguments);
        return ga('send', 'event', 'Log', str(arguments));
      },
      error: function() {
        console.error.apply(console, arguments);
        return ga('send', 'event', 'Error', str(arguments));
      }
    };
  });

  App.factory('key', function() {
    var key;
    key = Object.create(Mousetrap);
    key.bind = function(keys, fn) {
      var k, newkeys, _i, _len;
      if (!(keys instanceof Array)) {
        keys = [keys];
      }
      newkeys = [];
      for (_i = 0, _len = keys.length; _i < _len; _i++) {
        k = keys[_i];
        if (/both/.test(k)) {
          newkeys.push(k.replace("both", "ctrl"));
          newkeys.push(k.replace("both", "command"));
        } else {
          newkeys.push(k);
        }
      }
      return Mousetrap.bind(newkeys, fn);
    };
    return key;
  });

  App.factory('storage', function() {
    var storage, wrap;
    wrap = function(ns, fn) {
      return function() {
        arguments[0] = [ns, arguments[0]].join('-');
        return fn.apply(null, arguments);
      };
    };
    storage = {
      create: function(ns) {
        var fn, k, s;
        s = {};
        for (k in storage) {
          fn = storage[k];
          s[k] = wrap(ns, fn);
        }
        return s;
      },
      get: function(key) {
        var str;
        str = localStorage.getItem(key);
        if (str && str.substr(0, 4) === "J$ON") {
          return JSON.parse(str.substr(4));
        }
        return str;
      },
      set: function(key, val) {
        if (typeof val === 'object') {
          val = "J$ON" + (JSON.stringify(val));
        }
        return localStorage.setItem(key, val);
      },
      del: function(key) {
        return localStorage.removeItem(key);
      }
    };
    return window.storage = storage;
  });

  App.factory('$exceptionHandler', function(console) {
    return function(exception, cause) {
      console.error('Exception caught\n', exception.stack || exception);
      if (cause) {
        return console.error('Exception cause', cause);
      }
    };
  });

  App.run(function($rootScope, console) {
    window.root = $rootScope;
    console.log('Init');
    return $("#loading-cover").fadeOut(300, function() {
      return $(this).remove();
    });
  });

}).call(this);
