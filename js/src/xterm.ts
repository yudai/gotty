import { Terminal as bare, IDisposable } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { lib } from "libapps"


export class Xterm {
    elem: HTMLElement;
    term: bare;
    fitAddon: FitAddon;
    resizeListener: () => void;
    dataDisposable: IDisposable | null = null;
    resizeDisposable: IDisposable | null = null;
    decoder: lib.UTF8Decoder;

    message: HTMLElement;
    messageTimeout: number;
    messageTimer: number;


    constructor(elem: HTMLElement) {
        this.elem = elem;
        this.term = new bare({
            fontFamily: '"DejaVu Sans Mono", "Everson Mono", FreeMono, Menlo, Terminal, monospace, "Apple Symbols"',
        });
        this.fitAddon = new FitAddon();
        this.term.loadAddon(this.fitAddon);

        this.message = elem.ownerDocument.createElement("div");
        this.message.className = "xterm-overlay";
        this.messageTimeout = 2000;

        this.resizeListener = () => {
            this.fitAddon.fit();
            this.term.scrollToBottom();
            this.showMessage(String(this.term.cols) + "x" + String(this.term.rows), this.messageTimeout);
        };

        this.term.open(elem);
        this.term.focus();
        this.resizeListener();
        window.addEventListener("resize", this.resizeListener);

        this.decoder = new lib.UTF8Decoder()
    };

    info(): { columns: number, rows: number } {
        return { columns: this.term.cols, rows: this.term.rows };
    };

    output(data: string) {
        this.term.write(this.decoder.decode(data));
    };

    showMessage(message: string, timeout: number) {
        this.message.textContent = message;
        this.elem.appendChild(this.message);

        if (this.messageTimer) {
            clearTimeout(this.messageTimer);
        }
        if (timeout > 0) {
            this.messageTimer = window.setTimeout(() => {
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
    };

    onInput(callback: (input: string) => void) {
        this.dataDisposable = this.term.onData((data) => {
            callback(data);
        });

    };

    onResize(callback: (colmuns: number, rows: number) => void) {
        this.resizeDisposable = this.term.onResize((data) => {
            callback(data.cols, data.rows);
        });
    };

    deactivate(): void {
        if (this.dataDisposable) {
            this.dataDisposable.dispose();
            this.dataDisposable = null;
        }
        if (this.resizeDisposable) {
            this.resizeDisposable.dispose();
            this.resizeDisposable = null;
        }
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
