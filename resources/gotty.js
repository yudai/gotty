(function() {
    var httpsEnabled = window.location.protocol == "https:";
    var url = (httpsEnabled ? 'wss://' : 'ws://') + window.location.host + window.location.pathname + 'ws';
    var protocols = ["gotty"];
    var autoReconnect = -1;

    var openWs = function() {
        var ws = new WebSocket(url, protocols);

        var term;

        ws.onopen = function(event) {
            hterm.defaultStorage = new lib.Storage.Local();
            hterm.defaultStorage.clear();

            term = new hterm.Terminal();

            term.getPrefs().set("send-encoding", "raw");

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
            data = event.data.slice(1);
            switch(event.data[0]) {
            case '0':
                term.io.writeUTF16(data);
                break;
            case '1':
                term.setWindowTitle(data);
                break;
            case '2':
                preferences = JSON.parse(data);
                Object.keys(preferences).forEach(function(key) {
                    term.getPrefs().set(key, preferences[key]);
                });
                break;
            case '3':
                autoReconnect = JSON.parse(data);
                break;
            }
        }

        ws.onclose = function(event) {
            if (term) {
                term.uninstallKeyboard();
                term.io.showOverlay("Connection Closed", null);
            }
            tryReconnect();
        }

        ws.onerror = function(error) {
            tryReconnect();
        }
    }

    openWs();

    var tryReconnect = function() {
        if (autoReconnect >= 0) {
            setTimeout(openWs, autoReconnect * 1000);
        }
    }
})()
