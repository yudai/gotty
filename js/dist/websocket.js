"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var ConnectionFactory = (function () {
    function ConnectionFactory(url, protocols) {
        this.url = url;
        this.protocols = protocols;
    }
    ;
    ConnectionFactory.prototype.create = function () {
        return new Connection(this.url, this.protocols);
    };
    ;
    return ConnectionFactory;
}());
exports.ConnectionFactory = ConnectionFactory;
var Connection = (function () {
    function Connection(url, protocols) {
        this.bare = new WebSocket(url, protocols);
    }
    Connection.prototype.open = function () {
        // nothing todo for websocket
    };
    ;
    Connection.prototype.close = function () {
        this.bare.close();
    };
    ;
    Connection.prototype.send = function (data) {
        this.bare.send(data);
    };
    ;
    Connection.prototype.isOpen = function () {
        if (this.bare.readyState == WebSocket.CONNECTING ||
            this.bare.readyState == WebSocket.OPEN) {
            return true;
        }
        return false;
    };
    Connection.prototype.onOpen = function (callback) {
        this.bare.onopen = function (event) {
            callback();
        };
    };
    ;
    Connection.prototype.onReceive = function (callback) {
        this.bare.onmessage = function (event) {
            callback(event.data);
        };
    };
    ;
    Connection.prototype.onClose = function (callback) {
        this.bare.onclose = function (event) {
            callback();
        };
    };
    ;
    return Connection;
}());
exports.Connection = Connection;
//# sourceMappingURL=websocket.js.map