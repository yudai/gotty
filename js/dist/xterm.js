"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var bare = require("xterm");
bare.loadAddon("fit");
var TermXterm = (function () {
    function TermXterm(elem) {
        var _this = this;
        this.elem = elem;
        this.term = new bare();
        this.message = elem.ownerDocument.createElement("div");
        this.message.className = "xterm-overlay";
        this.messageTimeout = 2000;
        this.resizeListener = function () {
            _this.term.fit();
            _this.term.scrollToBottom();
            _this.showMessage(String(_this.term.cols) + "x" + String(_this.term.rows), _this.messageTimeout);
        };
        this.term.on("open", function () {
            _this.term.fit();
            _this.term.scrollToBottom();
            window.addEventListener("resize", function () { _this.resizeListener(); });
        });
        this.term.open(elem, true);
    }
    ;
    TermXterm.prototype.info = function () {
        return { columns: this.term.cols, rows: this.term.rows };
    };
    ;
    TermXterm.prototype.output = function (data) {
        this.term.write(data);
    };
    ;
    TermXterm.prototype.showMessage = function (message, timeout) {
        var _this = this;
        this.message.textContent = message;
        this.elem.appendChild(this.message);
        if (this.messageTimer) {
            clearTimeout(this.messageTimer);
        }
        if (timeout > 0) {
            this.messageTimer = setTimeout(function () {
                _this.elem.removeChild(_this.message);
            }, timeout);
        }
    };
    ;
    TermXterm.prototype.removeMessage = function () {
        if (this.message.parentNode == this.elem) {
            this.elem.removeChild(this.message);
        }
    };
    TermXterm.prototype.setWindowTitle = function (title) {
        document.title = title;
    };
    ;
    TermXterm.prototype.setPreferences = function (value) {
    };
    ;
    TermXterm.prototype.onInput = function (callback) {
        this.term.on("data", function (data) {
            callback(data);
        });
    };
    ;
    TermXterm.prototype.onResize = function (callback) {
        this.term.on("resize", function (data) {
            callback(data.cols, data.rows);
        });
    };
    ;
    TermXterm.prototype.deactivate = function () {
        this.term.off("data");
        this.term.off("resize");
        this.term.blur();
    };
    TermXterm.prototype.reset = function () {
        this.removeMessage();
        this.term.clear();
    };
    TermXterm.prototype.close = function () {
        window.removeEventListener("resize", this.resizeListener);
        this.term.destroy();
    };
    return TermXterm;
}());
exports.TermXterm = TermXterm;
//# sourceMappingURL=xterm.js.map