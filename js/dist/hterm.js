"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var bare = require("hterm");
var TermHterm = (function () {
    function TermHterm(elem) {
        var _this = this;
        this.elem = elem;
        this.term = new bare.Terminal();
        //        this.term.defaultStorage = new lib.Storage.Memory();
        this.term.getPrefs().set("send-encoding", "raw");
        this.term.decorate(this.elem);
        this.term.onTerminalReady = function () {
            _this.io = _this.term.io.push();
            _this.term.installKeyboard();
        };
    }
    ;
    TermHterm.prototype.info = function () {
        return { columns: this.term.screen.getWidth(), rows: this.term.screen.getHeight() };
    };
    ;
    TermHterm.prototype.output = function (data) {
        if (this.term.io.writeUTF8 != null) {
            this.term.io.writeUTF8(data);
        }
    };
    ;
    TermHterm.prototype.showMessage = function (message, timeout) {
        this.term.io.showOverlay(message, timeout);
    };
    ;
    TermHterm.prototype.removeMessage = function () {
        this.term.io.showOverlay("", 0);
    };
    TermHterm.prototype.setWindowTitle = function (title) {
        this.term.setWindowTitle(title);
    };
    ;
    TermHterm.prototype.setPreferences = function (value) {
        var _this = this;
        Object.keys(value).forEach(function (key) {
            _this.term.getPrefs().set(key, value[key]);
        });
    };
    ;
    TermHterm.prototype.onInput = function (callback) {
        this.io.onVTKeystroke = function (data) {
            callback(data);
        };
        this.io.sendString = function (data) {
            callback(data);
        };
    };
    ;
    TermHterm.prototype.onResize = function (callback) {
        this.io.onTerminalResize = function (columns, rows) {
            callback(columns, rows);
        };
    };
    ;
    TermHterm.prototype.deactivate = function () {
        this.io.onVTKeystroke = null;
        this.io.sendString = null;
        this.io.onTerminalResize = null;
        this.term.uninstallKeyboard();
    };
    TermHterm.prototype.reset = function () {
        this.removeMessage();
        //        this.term.reset();
    };
    TermHterm.prototype.close = function () {
        this.term.uninstallKeyboard();
    };
    return TermHterm;
}());
exports.TermHterm = TermHterm;
//# sourceMappingURL=hterm.js.map