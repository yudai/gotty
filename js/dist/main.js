"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var hterm_1 = require("./hterm");
var xterm_1 = require("./xterm");
var webtty_1 = require("./webtty");
var websocket_1 = require("./websocket");
var elem = document.getElementById("terminal");
if (elem !== null) {
    var term_1 = new xterm_1.TermXterm(elem);
    var httpsEnabled = window.location.protocol == "https:";
    var url = (httpsEnabled ? 'wss://' : 'ws://') + window.location.hostname + ":8080/ws";
    var args = window.location.search;
    var factory = new websocket_1.ConnectionFactory(url, webtty_1.protocols);
    var wt = new webtty_1.WebTTY(term_1, factory, args, "");
    var closer_1 = wt.open();
    new hterm_1.TermHterm(elem);
    window.addEventListener("unload", function () {
        closer_1();
        term_1.close();
    });
}
;
//# sourceMappingURL=main.js.map