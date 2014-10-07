(function() {
  var App;

  App = window.App = angular.module('sandbox', []);

  App.controller('Controls', function($rootScope, $scope, $window, ace, storage, key, $http, render, console) {
    var checkHash, hash, loadShare, loc, scope;
    scope = $rootScope.controls = $scope;
    scope["super"] = /Mac|iPod|iPhone|iPad/.test(navigator.userAgent) ? "⌘" : "Ctrl";
    key.bind(['both+enter', 'shift+enter'], function() {
      return scope.compile();
    });
    key.bind(['both+\\'], function() {
      return scope.imports();
    });
    key.bind(['both+.'], function() {
      return scope.share();
    });
    scope.compile = function() {
      $rootScope.loading = true;
      return $http({
        method: 'POST',
        url: "/compile",
        data: $.param({
          version: 2,
          body: ace.get()
        }),
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }).then(function(resp) {
        console.log('compiled');
        if (resp.data.compile_errors) {
          render({
            Errors: resp.data.compile_errors
          });
        } else {
          render(resp.data);
        }
      })["catch"](function(err) {
        return console.error("compile failed, oh noes", err);
      })["finally"](function() {
        $rootScope.loading = false;
        return ace.readonly(false);
      });
    };
    scope.imports = function() {
      $rootScope.loading = true;
      ace.readonly(true);
      return $http.post("/imports", ace.get()).then(function(resp) {
        console.log('formatted');
        return ace.set(resp.data);
      })["catch"](function(resp) {
        if (resp.data) {
          return render({
            Errors: resp.data,
            Events: null
          });
        }
      })["finally"](function() {
        $rootScope.loading = false;
        return ace.readonly(false);
      });
    };
    loc = window.location;
    $rootScope.shareURL = null;
    scope.share = function() {
      $rootScope.loading = true;
      return $http.post("/share", ace.get()).then(function(resp) {
        console.log('shared');
        return $rootScope.shareURL = loc.protocol + "//" + loc.host + "/#/" + resp.data;
      })["catch"](function(resp) {
        return console.error("share failed, oh noes", resp);
      })["finally"](function() {
        return $rootScope.loading = false;
      });
    };
    loadShare = function(id) {
      ace.readonly(true);
      $rootScope.loading = true;
      return $http.get("/p/" + id).then(function(resp) {
        var str, ta;
        str = resp.data;
        ta = $(str.substring(str.indexOf("<textarea"), str.indexOf("</textarea>") + 12));
        return ace.set(ta.val());
      })["catch"](function(resp) {
        return console.error("load share failed, oh noes", resp);
      })["finally"](function() {
        ace.readonly(false);
        return $rootScope.loading = false;
      });
    };
    hash = null;
    checkHash = function() {
      if (hash === loc.hash) {
        return;
      }
      hash = loc.hash;
      if (!/#\/([\w-]+)$/.test(hash)) {
        return;
      }
      return loadShare(RegExp.$1);
    };
    return setInterval(checkHash, 1000);
  });

  App.directive('dragger', function(ace, storage) {
    return {
      restrict: 'C',
      link: function(scope, element) {
        var dragging, init, resize;
        resize = function(percent) {
          var bot, top;
          top = percent + "%";
          bot = (100 - percent) + "%";
          $("#ace").css("height", top);
          ace._editor.resize();
          $("#dragger").css("top", top);
          return $("#out").css("top", top).css("height", bot);
        };
        init = storage.get('dragger-percent');
        if (init) {
          resize(init);
        }
        dragging = false;
        element.on("mousedown", function(e) {
          dragging = true;
          e.preventDefault();
        });
        $(window).on("mousemove", function(e) {
          var percent;
          if (!dragging) {
            return;
          }
          percent = 100 * (e.pageY / $(window).height());
          storage.set('dragger-percent', percent);
          resize(percent);
        });
        $(window).on("mouseup", function() {
          if (!dragging) {
            return;
          }
          dragging = false;
        });
      }
    };
  });

  App.directive('sharebox', function(ace, storage) {
    return {
      restrict: 'C',
      link: function(scope, element) {
        var t, ta;
        window.sharebox = scope;
        ta = element.find("textarea");
        scope.$root.$watch('shareURL', function(url) {
          if (!url) {
            return;
          }
          element.fadeIn();
          return ta.val(url).focus().select();
        });
        t = 0;
        ta.on("focus", function() {
          return clearTimeout(t);
        });
        ta.on("blur", function() {
          t = setTimeout(function() {
            scope.$root.shareURL = null;
            return element.fadeOut();
          }, 2000);
        });
      }
    };
  });

  App.factory('ace', function($rootScope, storage, key) {
    var Range, Selection, editor, markers, scope, session, unhight;
    storage = storage.create('ace');
    scope = $rootScope.ace = $rootScope.$new(true);
    Range = ace.require('ace/range').Range;
    Selection = ace.require('ace/selection').Selection;
    window.Selection = Selection;
    editor = ace.edit("ace");
    session = editor.getSession();
    scope._ace = ace;
    scope._editor = editor;
    scope._session = session;
    editor.commands.on("beforeExec", function(e) {
      console.log('command', e);
      if (e.command.name === "del") {
        editor.execCommand("duplicateSelection");
        e.preventDefault();
        return;
      }
    });
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
      var c;
      c = editor.getCursorPosition();
      session.setValue(val);
      return root.ace._session.getSelection().moveCursorTo(c.row, c.column);
    };
    scope.readonly = function(val) {
      return editor.setReadOnly(!!val);
    };
    scope.get = function() {
      return session.getValue();
    };
    unhight = 0;
    markers = [];
    scope.highlight = function(loc) {
      var m, r;
      clearTimeout(unhight);
      r = new Range(loc.row, loc.col, loc.row, loc.col + 1);
      m = session.addMarker(r, "ace_warning", "text", true);
      return markers.push(m);
    };
    scope.unhighlight = function() {
      var m;
      clearTimeout(unhight);
      while (markers.length) {
        m = markers.pop();
        session.removeMarker(m);
      }
    };
    scope.config({
      theme: "chrome",
      mode: "golang",
      tabSize: 4,
      softTabs: false,
      printMargin: false
    });
    scope.set(storage.get('current-code') || "package main\n\nfunc main() {\n\tprintln(42)\n}");
    editor.on('change', function() {
      clearTimeout(unhight);
      unhight = setTimeout(scope.unhighlight, 1000);
      return storage.set('current-code', scope.get());
    });
    return scope;
  });

  App.factory('console', function() {
    var str;
    ga('create', 'UA-38709761-13', 'auto');
    ga('send', 'pageview');
    setInterval((function() {
      return ga('send', 'event', 'Ping');
    }), 60 * 1000);
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

  App.factory('render', function(ace) {
    var clear, contents, handleErrors, handleEvents, render, timer, write;
    contents = document.getElementById("contents");
    clear = function() {
      return contents.innerHTML = '';
    };
    write = function(type, msg) {
      var span;
      if (/\u000c([^\u000c]*)$/.test(msg)) {
        msg = RegExp.$1;
        clear();
      }
      span = document.createElement("span");
      span.className = type;
      if (type === "err") {
        msg += "\n";
      }
      $(span).text(msg);
      contents.appendChild(span);
    };
    handleErrors = function(errstr) {
      var col, err, errs, msg, row, _i, _len;
      errs = errstr.split("\n");
      errs.unshift();
      for (_i = 0, _len = errs.length; _i < _len; _i++) {
        err = errs[_i];
        if (!err) {
          continue;
        }
        if (/^prog\.go:(\d+):((\d+):)?\ (.+)$/.test(err)) {
          row = parseInt(RegExp.$1, 10) - 1;
          col = RegExp.$3;
          msg = RegExp.$4;
          ace.highlight({
            row: row,
            col: col
          });
        }
        write('err', err);
      }
    };
    timer = 0;
    handleEvents = function(events) {
      var e, ms, next;
      clearTimeout(timer);
      if (events.length === 0) {
        write("exit", "\nProgram exited.");
        return;
      }
      next = handleEvents.bind(null, events);
      e = events[0];
      if (e.Delay) {
        ms = e.Delay / 1000000;
        e.Delay = 0;
        timer = setTimeout(next, ms);
        return;
      }
      write('out', e.Message);
      events.shift();
      next();
    };
    render = function(data) {
      if (!data) {
        return;
      }
      clear();
      clearTimeout(timer);
      if (data.Errors) {
        handleErrors(data.Errors);
      }
      if (data.Events) {
        return handleEvents(data.Events);
      }
    };
    return render;
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
    $("#loading-cover").fadeOut(500, function() {
      return $(this).remove();
    });
  });

}).call(this);
