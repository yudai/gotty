/**
 * Waits for a DOM element matching the selector to exist in the document.
 * Resolves immediately if it already exists.
 *
 * @param selector CSS selector for the element to wait for
 * @param timeout Optional timeout in milliseconds (default: no timeout)
 * @returns Promise that resolves with the found element
 */
export function waitForElement<T extends Element>(
  selector: string,
  timeout?: number,
): Promise<T> {
  return new Promise((resolve, reject) => {
    // If it already exists, resolve immediately
    const existing = document.querySelector<T>(selector);
    if (existing) {
      resolve(existing);
      return;
    }

    const observer = new MutationObserver(() => {
      const el = document.querySelector<T>(selector);
      if (el) {
        observer.disconnect();
        resolve(el);
      }
    });

    observer.observe(document.body, { childList: true, subtree: true });

    if (timeout) {
      setTimeout(() => {
        observer.disconnect();
        reject(new Error(`Timeout waiting for element: ${selector}`));
      }, timeout);
    }
  });
}
