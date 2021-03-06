/**
 * @file [Please Input File Description]
 * @author leeight(leeight@gmail.com)
 */

define(function (require) {

    // Action配置
    // 如果期望添加action时工具自动配置，请保持actionsConfig名称不变
    var actionsConfig = [
        {
            type: 'mail/Search',
            path: '/mail/search'
        },
        {
            type: 'mail/Compose',
            path: '/mail/compose'
        },
        {
            type: 'mail/View',
            path: '/mail/view'
        },
        {
            type: 'mail/Inbox',
            path: '/mail/inbox'
        },
        {
            type: 'mail/Inbox',
            path: '/mail/starred'
        },
        {
            type: 'mail/Inbox',
            path: '/mail/sent'
        },
        {
            type: 'mail/Inbox',
            path: '/calendar/list'
        },
        {
            type: 'mail/Inbox',
            path: '/mail/deleted'
        }
    ];

    var controller = require('er/controller');
    controller.registerAction(actionsConfig);

    // 这里可以添加一些模块配置
    // 如请求地址，表格fields等
    // 国际化相关语言定义，请使用lang，不建议在config中定义
    var config = {
    };

    return config;
});
