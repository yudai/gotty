import { OurXterm } from "./xterm";
import { Terminal, WebTTY, protocols } from "./webtty";
import { ConnectionFactory } from "./websocket";

// @TODO remove these
declare var gotty_auth_token: string;
declare var gotty_term: string;

const elem = document.getElementById("terminal")

if (elem !== null) {
    var term: Terminal;
    term = new OurXterm(elem);

    const httpsEnabled = window.location.protocol == "https:";
    const url = (httpsEnabled ? 'wss://' : 'ws://') + window.location.host + window.location.pathname + 'ws';
    const args = window.location.search;
    const factory = new ConnectionFactory(url, protocols);
    const wt = new WebTTY(term, factory, args, gotty_auth_token);
    const closer = wt.open();

    // According to https://developer.mozilla.org/en-US/docs/Web/API/Window/unload_event
    // this event is unreliable and in some cases (Firefox is mentioned), having an
    // "unload" event handler can have unwanted side effects. Consider commenting it out.
    window.addEventListener("unload", () => {
        closer();
        term.close();
    });
};
