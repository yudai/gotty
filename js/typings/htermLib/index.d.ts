export interface Storage {
}

export interface Memory {
    new (): Storage;
    Memory(): Storage
}

export var Storage: {
    Memory: Memory
};
