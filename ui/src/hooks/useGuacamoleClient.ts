import { ref } from 'vue';
// @ts-ignore
// import Guacamole from 'guacamole-common-js';
import Guacamole from '@dushixiang/guacamole-common-js';

export function useGuacamoleClient() {
  const client = ref<Guacamole.Client | null>(null);
  const tunnel = ref<Guacamole.Tunnel | null>(null);
  const display = ref<Guacamole.Display | null>(null);
  const audio = ref<Guacamole.Audio | null>(null);
  const audioStream = ref<Guacamole.AudioStream | null>(null);
  const recorder = ref<Guacamole.Recorder | null>(null);
  const fsObject = ref<Guacamole.File | null>(null);
  const driverName = ref<string | null>(null);

  function connectToGuacamole(url: string) {}

  function disconnect() {}
  function sendKeyEvent(keyCode: number, pressed: boolean) {}
  function sendMouseEvent(button: number, x: number, y: number, pressed: boolean) {}
  function sendTouchEvent(touchId: number, x: number, y: number, pressed: boolean) {}
  function sendClipboardData(data: string) {}
}
