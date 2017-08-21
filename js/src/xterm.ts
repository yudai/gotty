import * as bare from "xterm";

bare.loadAddon("fit");

export class TermXterm {
    elem: HTMLElement;

    message: HTMLElement;
    messageTimeout: number;
    messageTimer: number;

    term: bare;
    resizeListener: () => void;

    constructor(elem: HTMLElement) {
        this.elem = elem;
        this.term = new bare();

        this.message = elem.ownerDocument.createElement("div");
        this.message.className = "xterm-overlay";
        this.messageTimeout = 2000;


        this.resizeListener = () => {
            this.term.fit();
            this.term.scrollToBottom();
            this.showMessage(String(this.term.cols) + "x" + String(this.term.rows), this.messageTimeout);
        };

        this.term.on("open", () => {
            this.resizeListener();
            window.addEventListener("resize", () => { this.resizeListener(); });
        });

        this.term.open(elem, true);
    };

    info(): { columns: number, rows: number } {
        return { columns: this.term.cols, rows: this.term.rows };
    };

    output(data: string) {
        this.term.write(data);
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
    };

    onInput(callback: (input: string) => void) {
        this.term.on("data", (data) => {
            callback(data);
        });

    };

    onResize(callback: (colmuns: number, rows: number) => void) {
        this.term.on("resize", (data) => {
            callback(data.cols, data.rows);
        });
    };

    deactivate(): void {
        this.term.off("data");
        this.term.off("resize");
        this.term.blur();
    }

    reset(): void {
        this.removeMessage();
        this.term.clear();
    }

    close(): void {
        window.removeEventListener("resize", this.resizeListener);
        this.term.destroy();
    }
}
