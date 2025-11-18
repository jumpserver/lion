import { useClipboardItems } from '@vueuse/core';
const { content, read } = useClipboardItems();

import { effect, shallowRef } from 'vue';
const computedText = shallowRef('');
const computedMimeType = shallowRef('');
effect(() => {
  Promise.all(content.value.map((item) => item.getType('text/plain'))).then(async (blobs) => {
    computedMimeType.value = blobs.map((blob) => blob.type).join(', ');
    computedText.value = (await Promise.all(blobs.map((blob) => blob.text()))).join(', ');
  });
});
export async function readClipboardText(): Promise<string> {
  try {
    await read();
    return computedText.value;
  } catch (err) {
    console.error('Failed to read clipboard:', err);
    return '';
  }
}
