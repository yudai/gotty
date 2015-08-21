(function() {
    var httpsEnabled = window.location.protocol == "https:";
    var url = (httpsEnabled ? 'wss://' : 'ws://') + window.location.host + window.location.pathname + 'ws';
    var protocols = ["gotty"];
    var ws = new WebSocket(url, protocols);

    var term;

    ws.onopen = function(event) {
        hterm.defaultStorage = new lib.Storage.Local()

        term = new hterm.Terminal();

        term.onTerminalReady = function() {
            var io = term.io.push();

            io.onVTKeystroke = function(str) {
                ws.send("0" + str);
            };

            io.sendString = io.onVTKeystroke;

            io.onTerminalResize = function(columns, rows) {
                ws.send(
                    "1" + JSON.stringify(
                        {
                            columns: columns,
                            rows: rows,
                        }
                    )
                )
            };

            term.installKeyboard();
        };

        term.decorate(document.body);
    };

    ws.onmessage = function(event) {
        term.io.writeUTF16(event.data);
    }

    ws.onclose = function(event) {
        term.io.showOverlay("Connection Closed", null);
        term.uninstallKeyboard();
    }
})()
