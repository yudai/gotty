/**
 * Waits for a DOM element matching the selector to exist in the document.
 * Resolves immediately if it already exists.
 *
 * @param selector CSS selector for the element to wait for
 * @param timeout Optional timeout in milliseconds (default: no timeout)
 * @returns Promise that resolves with the found element
 */
export declare function waitForElement<T extends Element>(selector: string, timeout?: number): Promise<T>;
