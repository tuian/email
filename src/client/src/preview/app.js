/**
 * Office文档的预览页面，现在仅支持xlsx的格式
 * 预览的时候一些样式（前景色，背景色）可能会被丢弃掉
 */
define(function(require) {
var ajax = require('er/ajax');
var URL = require('er/URL');
var lib = require('esui/lib');

var exports = {};

function failure() {
    lib.g('main').innerHTML = 'FAILURE';
}

var gSheets = null;
var gFile = URL.parseQuery(location.search.substr(1)).file;

function buildSheets(sheets, opt_idx) {
    var idx = opt_idx || 0;
    var max = 0;
    var sheet = sheets[idx];
    for (var i = 0; i < sheet.data.length; i ++) {
        max = Math.max(max, sheet.data[i].length);
    }

    var navs = ['<ul class="xlsx-navs">'];
    for (var i = 0; i < sheets.length; i ++) {
        navs.push('<li ' +
            'data-index="' + i + '"' +
            ((i === idx) ? ' class="active"' : '') + '>' +
            (sheets[i].name || '(No name)') + '</li>');
    }
    navs.push('</ul>');

    var html = [
        '<div class="xlsx-container">',
        '<div class="xlsx-inner">',
        '<div class="xlsx">',
        '<h2>', gFile, '</h2>',
        navs.join(''),
        '<table class="xlsx-table" width="' + (max * 100 + 30) + 'px" cellpadding="5" cellspacing="0" border="1">'
    ];

    // colgroup
    html.push('<colgroup>');
    for (var j = 0; j < max; j ++) {
        html.push('<col />');
    }
    html.push('</colgroup>')

    // thead
    html.push('<tr>');
    html.push('<th>&nbsp;</th>');
    for (var j = 0; j < max; j ++) {
        html.push('<th>' + String.fromCharCode(65 + j) + '</th>');
    }
    html.push('</tr>');

    for (var i = 0; i < sheet.data.length; i ++) {
        var row = sheet.data[i];
        html.push('<tr><th>' + (i + 1) + '</th>');
        for (var j = 0; j < max; j ++) {
            var text = (row[j] || '&nbsp;');
            html.push('<td title="' + text + '">' + text.replace(/\r?\n/g, '<br/>') + '</td>');
        }
        html.push('</tr>');
    }
    html.push('</table></div></div></div>');

    lib.g('main').innerHTML = html.join('');
}

function initEvents() {
    lib.g('main').onclick = function(evt) {
        var target = evt.target || evt.srcElement;
        if (target.nodeType === 1 && target.nodeName === 'LI') {
            var idx = parseInt(target.getAttribute('data-index'), 10);
            buildSheets(gSheets, idx);
        }
    }
}

exports.start = function() {
    var apiUrl = '/api/doc/preview' + location.search;
    ajax.getJSON(apiUrl).then(function(data) {
        if (/\.docx$/.test(gFile)) {
            lib.g('main').innerHTML =
              '<div class="docx-container"><div class="docx-inner"><div class="docx">' +
                '<h2>' + gFile + '</h2>' +
                data.result + '</div></div></div>';
        }
        else if (/\.xlsx$/.test(gFile)) {
            gSheets = data.result;
            buildSheets(gSheets, 0);
            initEvents();
        }
        else {
            alert('Invalid doc format.');
        }
    }, failure);
};

return exports;
});
