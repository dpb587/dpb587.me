define(
  [ 'vendor/mootools' ],
  function () {
      var _ = {};

      _.getCardType = function (num) {
          if (num.substring(0, 1) == '4') {
              return 'visa';
          } else if ((num.length == 16) && ((num.substring(0, 2).toInt() >= 51) && (num.substring(0, 2).toInt() <= 55))) {
              return 'mastercard';
          } else if ((num.length == 15) && (num.substring(0, 2) == '34') || (num.substring(0, 2) == '37')) {
              return 'amex';
          } else if ((num.length == 16) && (num.substring(0, 4) == '6011')) {
              return 'discover';
          }

          return null;
      }

      _.listen = function (dom, cb) {
          var dom = $(dom);
          var timer;

          function deferred() {
              var result = _.parse(dom.value);
  
              if (result) {
                  dom.value = '';
  
                  cb(result, dom);
              }
          }

          dom.addEvent(
              'keypress',
              function (e) {
                  if ((e.key == 'enter') && (dom.value.substring(0, 2) == '%B')) {
                      e.stop();

                      timer = setTimeout(deferred, 50);
                  } else if (timer) {
                      e.preventDefault();

                      clearTimeout(timer);
                      timer = setTimeout(deferred, 50);
                  }
              }
          );

          return _;
      };

      _.parse = function (data) {
          var result = {};

          if (data.substring(0, 2) != '%B') {
              return false;
          }

          var lines = data.split('\n');
          var line1 = lines[0].split('^');

          if (line1.length < 3) {
              return false;
          }

          result.number = line1[0].substring(2);

          var nameslash = line1[1].indexOf('/');

          if (-1 != nameslash) {
              result.name = line1[1].substring(nameslash + 1).trim() + ' ' + line1[1].substring(0, nameslash).trim();
          } else {
              result.name = line1[1].trim();
          }

          result.expm = line1[2].substring(2, 4);
          result.expy = '20' + line1[2].substring(0, 2);

          return result;
      };

      return _;
  }
);
