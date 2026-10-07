import { Terminal as bare, IDisposable } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { lib } from "libapps";
export declare class Xterm {
    elem: HTMLElement;
    term: bare;
    fitAddon: FitAddon;
    resizeListener: () => void;
    dataDisposable: IDisposable | null;
    resizeDisposable: IDisposable | null;
    decoder: lib.UTF8Decoder;
    message: HTMLElement;
    messageTimeout: number;
    messageTimer: number;
    constructor(elem: HTMLElement);
    info(): {
        columns: number;
        rows: number;
    };
    output(data: string): void;
    showMessage(message: string, timeout: number): void;
    removeMessage(): void;
    setWindowTitle(title: string): void;
    setPreferences(value: object): void;
    onInput(callback: (input: string) => void): void;
    onResize(callback: (colmuns: number, rows: number) => void): void;
    deactivate(): void;
    reset(): void;
    close(): void;
}
