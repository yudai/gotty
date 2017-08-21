"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.protocols = ["webtty"];
exports.msgInputUnknown = '0';
exports.msgInput = '1';
exports.msgPing = '2';
exports.msgResizeTerminal = '3';
exports.msgUnknownOutput = '0';
exports.msgOutput = '1';
exports.msgPong = '2';
exports.msgSetWindowTitle = '3';
exports.msgSetPreferences = '4';
exports.msgSetReconnect = '5';
var WebTTY = (function () {
    function WebTTY(term, connectionFactory, args, authToken) {
        this.term = term;
        this.connectionFactory = connectionFactory;
        this.args = args;
        this.authToken = authToken;
        this.reconnect = -1;
    }
    ;
    WebTTY.prototype.open = function () {
        var _this = this;
        var connection = this.connectionFactory.create();
        var pingTimer;
        var reconnectTimeout;
        var setup = function () {
            connection.onOpen(function () {
                var termInfo = _this.term.info();
                connection.send(JSON.stringify({
                    Arguments: _this.args,
                    AuthToken: _this.authToken,
                }));
                var resizeHandler = function (colmuns, rows) {
                    connection.send(exports.msgResizeTerminal + JSON.stringify({
                        columns: colmuns,
                        rows: rows
                    }));
                };
                _this.term.onResize(resizeHandler);
                resizeHandler(termInfo.columns, termInfo.rows);
                _this.term.onInput(function (input) {
                    connection.send(exports.msgInput + input);
                });
                pingTimer = setInterval(function () {
                    connection.send(exports.msgPing);
                }, 30 * 1000);
            });
            connection.onReceive(function (data) {
                var payload = data.slice(1);
                switch (data[0]) {
                    case exports.msgOutput:
                        _this.term.output(decodeURIComponent(Array.prototype.map.call(atob(payload), function (c) {
                            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
                        }).join('')));
                        break;
                    case exports.msgPong:
                        break;
                    case exports.msgSetWindowTitle:
                        _this.term.setWindowTitle(payload);
                        break;
                    case exports.msgSetPreferences:
                        var preferences = JSON.parse(payload);
                        _this.term.setPreferences(preferences);
                        break;
                    case exports.msgSetReconnect:
                        var autoReconnect = JSON.parse(payload);
                        console.log("Enabling reconnect: " + autoReconnect + " seconds");
                        _this.reconnect = autoReconnect;
                        break;
                }
            });
            connection.onClose(function () {
                clearInterval(pingTimer);
                _this.term.deactivate();
                _this.term.showMessage("Connection Closed", 0);
                if (_this.reconnect > 0) {
                    reconnectTimeout = setTimeout(function () {
                        connection = _this.connectionFactory.create();
                        _this.term.reset();
                        setup();
                    }, _this.reconnect * 1000);
                }
            });
            connection.open();
        };
        setup();
        return function () {
            clearTimeout(reconnectTimeout);
            connection.close();
        };
    };
    ;
    return WebTTY;
}());
exports.WebTTY = WebTTY;
;
//# sourceMappingURL=webtty.js.map