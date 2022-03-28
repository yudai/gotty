import { Terminal, IDisposable } from "xterm";
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { WebglAddon } from 'xterm-addon-webgl';

export class Xterm {
    elem: HTMLElement;
    term: Terminal;
    resizeListener: () => void;

    message: HTMLElement;
    messageTimeout: number;
    messageTimer: NodeJS.Timeout;
    onResizeHandler: IDisposable;
    onDataHandler: IDisposable;
    fitAddOn: FitAddon;

    constructor(elem: HTMLElement) {
        this.elem = elem;
        this.term = new Terminal();
        this.fitAddOn = new FitAddon();
        this.term.loadAddon(new WebLinksAddon());
        this.term.loadAddon(this.fitAddOn);

        this.message = elem.ownerDocument.createElement("div");
        this.message.className = "xterm-overlay";
        this.messageTimeout = 2000;

        this.resizeListener = () => {
            this.fitAddOn.fit();
            this.term.scrollToBottom();
            this.showMessage(String(this.term.cols) + "x" + String(this.term.rows), this.messageTimeout);
        };

        this.term.open(elem);
        this.term.focus();
        this.resizeListener();
        window.addEventListener("resize", () => { this.resizeListener(); });
    };

    info(): { columns: number, rows: number } {
        return { columns: this.term.cols, rows: this.term.rows };
    };

    output(data: string) {
        this.term.write(Uint8Array.from(data, c => c.charCodeAt(0)));
    };

    showMessage(message: string, timeout: number) {
        this.message.textContent = message;
        this.elem.appendChild(this.message);

        if (this.messageTimer) {
            clearTimeout(this.messageTimer);
        }
        if (timeout > 0) {
            this.messageTimer = setTimeout(() => {
                this.elem.removeChild(this.message);
            }, timeout);
        }
    };

    removeMessage(): void {
        if (this.message.parentNode == this.elem) {
            this.elem.removeChild(this.message);
        }
    }

    setWindowTitle(title: string) {
        document.title = title;
    };

    setPreferences(value: object) {
        Object.keys(value).forEach((key) => {
            if (key == "EnableWebGL" && key) {
                this.term.loadAddon(new WebglAddon());
            } else if (key == "font-size") {
                this.term.setOption("fontSize", value[key])
            } else if (key == "font-family") {
                this.term.setOption("fontFamily", value[key])
            }
        });
    };

    onInput(callback: (input: string) => void) {
        this.onDataHandler = this.term.onData((data) => {
            callback(data);
        });

    };

    onResize(callback: (colmuns: number, rows: number) => void) {
        this.onResizeHandler = this.term.onResize(() => {
            callback(this.term.cols, this.term.rows);
        });
    };

    deactivate(): void {
        this.onDataHandler.dispose();
        this.onResizeHandler.dispose();
        this.term.blur();
    }

    reset(): void {
        this.removeMessage();
        this.term.clear();
    }

    close(): void {
        window.removeEventListener("resize", this.resizeListener);
        this.term.dispose();
    }
}
