export function copyToClipboard(text: string): Promise<void> {
  return new Promise((resolve, reject) => {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard
        .writeText(text)
        .then(() => resolve())
        .catch((err) => reject(err));
    } else {
      // Fallback for browsers that do not support the Clipboard API
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed'; // Prevent scrolling to bottom of page in MS Edge.
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        document.execCommand('copy');
        resolve();
      } catch (err) {
        reject(err);
      } finally {
        document.body.removeChild(textArea);
      }
    }
  });
}

export function pasteFromClipboard(): Promise<string> {
  return new Promise((resolve, reject) => {
    if (navigator.clipboard && navigator.clipboard.readText) {
      navigator.clipboard
        .readText()
        .then((text) => resolve(text))
        .catch((err) => {
          console.error('Failed to read from  Clipboard API :', err);
          // reject(err);
          resolve(''); // Return empty string on error
        });
    } else {
      // Fallback for browsers that do not support the Clipboard API
      // 保存当前焦点元素
      const activeElement = document.activeElement as HTMLElement;
      const textArea = document.createElement('textarea');
      textArea.style.position = 'fixed'; // Prevent scrolling to bottom of page in MS Edge.
      textArea.style.opacity = '0'; // 使其不可见
      textArea.style.pointerEvents = 'none'; // 防止意外交互
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      try {
        const text = document.execCommand('paste') ? textArea.value : '';
        resolve(text);
      } catch (err) {
        console.error('Failed to read from clipboard:', err);
        // reject(err);
        resolve(''); // Return empty string on error
      } finally {
        document.body.removeChild(textArea);
        // 恢复原来的焦点
        if (activeElement && activeElement.focus) {
          activeElement.focus();
        }
      }
    }
  });
}

export async function writeToClipboard(text: string): Promise<void> {
  try {
    await copyToClipboard(text);
  } catch (err) {
    console.error('Failed to write text to clipboard:', err);
  }
}

export async function readClipboardText(): Promise<string> {
  return await pasteFromClipboard();
}
