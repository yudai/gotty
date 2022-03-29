import * as Zmodem from 'zmodem.js/src/zmodem_browser';

export const protocols = ["webtty"];

export const msgInputUnknown = '0';
export const msgInput = '1';
export const msgPing = '2';
export const msgResizeTerminal = '3';

export const msgUnknownOutput = '0';
export const msgOutput = '1';
export const msgPong = '2';
export const msgSetWindowTitle = '3';
export const msgSetPreferences = '4';
export const msgSetReconnect = '5';
export const msgSetBufferSize = '6';


export interface Terminal {
    info(): { columns: number, rows: number };
    output(data: string): void;
    showMessage(message: string, timeout: number): void;
    getMessage(): HTMLElement;
    removeMessage(): void;
    setWindowTitle(title: string): void;
    setPreferences(value: object): void;
    onInput(callback: (input: string) => void): void;
    onResize(callback: (colmuns: number, rows: number) => void): void;
    reset(): void;
    deactivate(): void;
    close(): void;
}

export interface Connection {
    open(): void;
    close(): void;
    send(data: string): void;
    isOpen(): boolean;
    onOpen(callback: () => void): void;
    onReceive(callback: (data: string) => void): void;
    onClose(callback: () => void): void;
}

export interface ConnectionFactory {
    create(): Connection;
}


export class WebTTY {
    term: Terminal;
    connectionFactory: ConnectionFactory;
    connection: Connection;
    args: string;
    authToken: string;
    reconnect: number;
    bufSize: number;
    sentry: Zmodem.Sentry;

    constructor(term: Terminal, connectionFactory: ConnectionFactory, args: string, authToken: string) {
        this.term = term;
        this.connectionFactory = connectionFactory;
        this.args = args;
        this.authToken = authToken;
        this.reconnect = -1;
        this.bufSize = 1024;

        this.sentry = new Zmodem.Sentry({
            'to_terminal': (d: any) => this.term.output(d),
            'on_detect': (detection: Zmodem.Detection) => this.zmodemDetect(detection),
            'sender': (x: Uint8Array) => this.sendInput(x),
            'on_retract': (x: any) => alert("never mind!"),
        })
    };

    private zmodemDetect(detection: Zmodem.Detection) {
        var zsession = detection.confirm();

        if (zsession.type === "send") {
            this.zmodemSend(zsession);
        }
        else {
            zsession.on("offer", (xfer: any) => this.zmodemOffer(xfer));
            zsession.start();
        }
    }

    private zmodemSend(zsession: any) {
        let dialog = this.getFileSendDialog();
        dialog.style.display = 'block';

        let selector = document.getElementById("sendFileSelector");
        if (selector != null) {
            selector.onchange = (event) => {
                Zmodem.Browser.send_files(zsession, (event.target as HTMLInputElement).files)
                    .then(() => zsession.close())
                    .catch(e => console.log(e));
                dialog.style.display = 'none';
            };
        }
    }

    private zmodemOffer(xfer: Zmodem.Offer) {
        var dialog = this.getFileAcceptanceDialog();
        dialog.style.display = 'block';

        var filenameElem = document.getElementById("filename");
        if (filenameElem != null) {
            filenameElem.textContent = xfer.get_details().name;
        }
        var sizeElem = document.getElementById("filesize");
        if (sizeElem != null) {
            sizeElem.textContent = xfer.get_details().size;
        }
        var skipLink = document.getElementById("skipTransfer");
        if (skipLink != null) {
            skipLink.onclick = (ev) => {
                xfer.skip();
                dialog.style.display = 'none';
            }
        }

        var acceptLink = document.getElementById("acceptTransfer");
        if (acceptLink != null) {
            acceptLink.onclick = (ev) => {
                dialog.style.display = 'none';
                xfer.accept().then((payloads: any) => {
                    //Now you need some mechanism to save the file.
                    //An example of how you can do this in a browser:
                    Zmodem.Browser.save_to_disk(
                        payloads,
                        xfer.get_details().name
                    );
                });
            }
        }
    }

    private sendInput(input: string | Uint8Array) {
        let effectiveBufferSize = this.bufSize - 1;
        let dataString: string

        if (Array.isArray(input)) {
            dataString = String.fromCharCode.apply(null, input);
        } else {
            dataString = (input as string);
        }

        // Account for base64 encoding
        let maxChunkSize = Math.floor(effectiveBufferSize / 4)*3;

        for (let i = 0; i < Math.ceil(dataString.length / maxChunkSize); i++) {
            let inputChunk = dataString.substring(i * effectiveBufferSize, Math.min((i + 1) * effectiveBufferSize, dataString.length))
            this.connection.send(msgInput + btoa(inputChunk));
        }
    }

    getFileAcceptanceDialog(): HTMLElement {
        let dialog = document.getElementById("acceptFileDialog");
        if (dialog == null) {
            dialog = document.createElement("div");
            dialog.id = 'acceptFileDialog';
            dialog.className = 'fileDialog';
            dialog.innerHTML = '<p>Incoming file transfer: <tt id="filename"></tt> (<span id="filesize"></span> bytes)</p><a id="acceptTransfer" href="#">Accept</a> <a id="skipTransfer" href="#">Decline</a>';
            document.body.appendChild(dialog);
        }
        return dialog;
    }

    getFileSendDialog(): HTMLElement {
        let dialog = document.getElementById("sendFileDialog");
        if (dialog == null) {
            dialog = document.createElement("div");
            dialog.id = 'sendFileDialog';
            dialog.className = 'fileDialog';
            dialog.innerHTML = '<p>Remote ready to receive files. <input id="sendFileSelector" class="file-input" type="file" multiple="" /></p>';
            document.body.appendChild(dialog);
        }
        return dialog;
    }

    open() {
        let connection = this.connectionFactory.create();
        let pingTimer: NodeJS.Timeout;
        let reconnectTimeout: NodeJS.Timeout;
        this.connection = connection;

        const setup = () => {
            connection.onOpen(() => {
                const termInfo = this.term.info();

                connection.send(JSON.stringify(
                    {
                        Arguments: this.args,
                        AuthToken: this.authToken,
                    }
                ));


                const resizeHandler = (colmuns: number, rows: number) => {
                    connection.send(
                        msgResizeTerminal + JSON.stringify(
                            {
                                columns: colmuns,
                                rows: rows
                            }
                        )
                    );
                };

                this.term.onResize(resizeHandler);
                resizeHandler(termInfo.columns, termInfo.rows);

                this.term.onInput(
                    (input: string) => {
                        this.sendInput(input);
                    }
                );

                pingTimer = setInterval(() => {
                    connection.send(msgPing)
                }, 30 * 1000);

            });

            connection.onReceive((data) => {
                const payload = data.slice(1);
                switch (data[0]) {
                    case msgOutput:
                        this.sentry.consume(Uint8Array.from(atob(payload), c => c.charCodeAt(0)));
                        break;
                    case msgPong:
                        break;
                    case msgSetWindowTitle:
                        this.term.setWindowTitle(payload);
                        break;
                    case msgSetPreferences:
                        const preferences = JSON.parse(payload);
                        this.term.setPreferences(preferences);
                        break;
                    case msgSetReconnect:
                        const autoReconnect = JSON.parse(payload);
                        console.log("Enabling reconnect: " + autoReconnect + " seconds")
                        this.reconnect = autoReconnect;
                        break;
                    case msgSetBufferSize:
                        const bufSize = JSON.parse(payload);
                        this.bufSize = bufSize;
                        break;
                }
            });

            connection.onClose(() => {
                clearInterval(pingTimer);
                this.term.deactivate();
                this.term.showMessage("Connection Closed", 0);
                if (this.reconnect > 0) {
                    reconnectTimeout = setTimeout(() => {
                        connection = this.connectionFactory.create();
                        this.term.reset();
                        setup();
                    }, this.reconnect * 1000);
                }
            });

            connection.open();
        }

        setup();
        return () => {
            clearTimeout(reconnectTimeout);
            connection.close();
        }
    };
};
