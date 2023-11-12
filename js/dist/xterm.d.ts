/// <reference types="node" />
import { lib } from "libapps";
import { IDisposable, Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
export declare class Xterm {
    elem: HTMLElement;
    term: Terminal;
    resizeListener: () => void;
    decoder: lib.UTF8Decoder;
    message: HTMLElement;
    messageTimeout: number;
    messageTimer: NodeJS.Timer;
    fitAddon: FitAddon;
    disposables: IDisposable[];
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
