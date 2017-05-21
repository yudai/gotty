export interface Terminal {
    io: IO;
    onTerminalReady: () => void;

    getPrefs(): Prefs;
    decorate(HTMLElement);
    installKeyboard(): void;
    uninstallKeyboard(): void;
    setWindowTitle(title: string): void;
    reset(): void;
    softReset(): void;
}

export interface TerminalConstructor {
    new (): Terminal;
    (): Terminal;
}


export interface IO {
    writeUTF8: ((data: string) => void);
    writeUTF16: ((data: string) => void);
    onVTKeystroke: ((data: string) => void) | null;
    sendString: ((data: string) => void) | null;
    onTerminalResize: ((columns: number, rows: number) => void) | null;

    push(): IO;
    writeUTF(data: string);
    showOverlay(message: string, timeout: number | null);
}

export interface Prefs {
    set(key: string, value: string): void;
}

export interface Storage {
}

export var Terminal: TerminalConstructor;

// @TODO: is there better way?
// exported variables are forced to be read-protected.
declare global {
    var hterm: {
        defaultStorage: Storage;
    };
}
