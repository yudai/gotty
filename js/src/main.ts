import { Xterm } from "./xterm";
import { Terminal, WebTTY, protocols } from "./webtty";
import { ConnectionFactory } from "./websocket";

const elem = document.getElementById("terminal")

if (elem !== null) {
    var term: Terminal = new Xterm(elem);
    const httpsEnabled = window.location.protocol == "https:";
    const url = (httpsEnabled ? 'wss://' : 'ws://') + window.location.host + window.location.pathname + 'ws';
    const factory = new ConnectionFactory(url, protocols);
    const wt = new WebTTY(term, factory);
    const closer = wt.open();

    window.addEventListener("unload", () => {
        closer();
        term.close();
    });
};
