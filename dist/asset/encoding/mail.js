define('encoding/mail', [
    'require',
    'underscore',
    './base64'
], function (require) {
    var u = require('underscore');
    var base64 = require('./base64');
    var exports = {};
    exports.encodeAddress = function (item) {
        if (!item.name) {
            return '<' + item.address + '>';
        } else {
            return '=?utf-8?b?' + base64.encode(item.name) + '?= <' + item.address + '>';
        }
    };
    exports.dumpAddress = function (addresses) {
        var dump = [];
        u.each(addresses || [], function (item) {
            if (item.name) {
                dump.push('"' + item.name + '"' + ' &lt;' + item.address + '&gt;');
            } else {
                dump.push(item.address);
            }
        });
        return dump.join(', ');
    };
    return exports;
});