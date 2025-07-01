<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue';
import { set, useWindowSize } from '@vueuse/core';
import { useDebounceFn } from '@vueuse/core';
// @ts-ignore
// import Guacamole from 'guacamole-common-js';
import Guacamole from '@dushixiang/guacamole-common-js';
import { NSpin, useMessage, NTabPane } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { getCurrentConnectParams } from '@/utils/common';
import { getSupportedGuacMimeTypes } from '@/utils/guacamole_helper';
import type { GuacamoleDisplay } from '@/types/guacamole.type';
import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';
import { ErrorStatusCodes, ConvertGuacamoleError } from '@/utils/status';
import { LanguageCode } from '@/locales';
import ClipBoardText from '@/components/ClipBoardText.vue';
import FileManager from '@/components/FileManager.vue';
import { readClipboardText } from '@/utils/clipboard';
const message = useMessage();
const { t } = useI18n();
const { width, height } = useWindowSize();
interface GuacamoleClient {
  sendSize: (width: number, height: number) => void;
  connect: (connectString: string) => void;
  getDisplay: () => GuacamoleDisplay;
  disconnect(): () => void;
  createClipboardStream: (type: string) => any;
}

const apiPrefix = ref('');
const wsPrefix = ref('');
const localCursor = ref(false);
const scale = ref(1);
const pixelDensity = window.devicePixelRatio || 1;
const sessionObject = ref();
const shareId = ref<string | null>(null);
const currentUser = ref<any | null>(null);
const onlineUsersMap = ref<Record<string, any>>({});
const warningIntervalId = ref<number | null>(null);
const action_permission = ref<any>();
const enableShare = ref(false);
const guacDisplay = ref<GuacamoleDisplay>();
const guacTunnel = ref(null);
const guacClient = ref<GuacamoleClient>();
const drawShow = ref(false);
const connectStatus = ref('Connecting');
const loading = ref(true);

const remoteClipboardText = ref<string>('');
const hasClipboardPermission = ref(false);
const enableFilesystem = ref(false);
// const keyboard = ref<Guacamole.Keyboard | null>(null);
const debouncedResize = useDebounceFn(() => {
  updateScale();
  if (guacClient.value && guacDisplay.value) {
    console.log('Sending resize to Guacamole client:', width.value, height.value);
    guacClient.value.sendSize(width.value, height.value);
  }
}, 300);

const updateScale = () => {
  if (!guacDisplay.value || !guacClient.value) {
    console.warn('Guacamole display is not initialized yet.');
    return;
  }
  const w = guacDisplay.value.getWidth();
  const h = guacDisplay.value.getHeight();

  if (h === 0 || w === 0) {
    return 1;
  }
  const newScale = Math.min(width.value / w, height.value / h);
  if (newScale !== scale.value) {
    console.log(`Guacamole display scaled from ${scale.value} to ${newScale}`);
    scale.value = newScale;
    guacDisplay.value.scale(newScale);
  }
};

watch(
  [width, height],
  ([newWidth, newHeight]) => {
    debouncedResize();
  },
  { immediate: true },
);

const getConnectString = async (tokenId: string) => {
  const optimalWidth = width.value;
  const optimalHeight = height.value;
  const optimalDpi = pixelDensity * 96;
  const supportMimeTypes = await getSupportedGuacMimeTypes();
  const connectString =
    'TOKEN_ID=' +
    encodeURIComponent(tokenId) +
    '&GUAC_WIDTH=' +
    Math.floor(optimalWidth) +
    '&GUAC_HEIGHT=' +
    Math.floor(optimalHeight) +
    '&GUAC_DPI=' +
    Math.floor(optimalDpi) +
    supportMimeTypes;
  return connectString;
};

const clientStateChanged = (state: any) => {
  console.log('Guacamole client state changed:', state);

  switch (state) {
    case 0:
      connectStatus.value = 'IDLE';
      break;
    case 1:
      connectStatus.value = 'Connecting';
      break;
    case 2:
      connectStatus.value = 'Connected + waiting';
      break;
    case 3:
      loading.value = false;
      connectStatus.value = 'Connected';
      break;
    case 4:
      connectStatus.value = 'Disconnecting';
      break;
    case 5:
      connectStatus.value = 'Disconnected';
      lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.CLOSE, '');
      message.error(t('Connection disconnected'));
      guacDisplay.value?.getElement()?.remove();
      break;
  }
};

const onJmsEvent = (event: any, data: any) => {
  console.log('Received JMS event:', event);
  const dataObj = JSON.parse(data);
  switch (event) {
    case 'session_pause': {
      const msg = `${dataObj.user} ${t('PauseSession')}`;
      message.info(msg);
      break;
    }
    case 'session_resume': {
      const msg = `${dataObj.user} ${t('ResumeSession')}`;
      message.info(msg);
      break;
    }
    case 'session': {
      sessionObject.value = dataObj;
      const action = dataObj.action_permission || {};
      action_permission.value = dataObj.action_permission || {};
      enableShare.value = action_permission.value.enable_share || false;
      hasClipboardPermission.value = action.enable_copy || action.enable_paste;
      console.log('Session object hasClipboardPermission:', hasClipboardPermission);

      break;
    }
    case 'current_user': {
      currentUser.value = dataObj;
      shareId.value = dataObj.share_id || null;
      break;
    }
    case 'share_join': {
      if (dataObj.primary) {
        break;
      }
      const joinMsg = `${dataObj.user} ${t('JoinShare')}`;
      message.info(joinMsg);
      break;
    }
    case 'share_exit': {
      const leaveMsg = `${dataObj.user} ${t('LeaveShare')}`;
      message.info(leaveMsg);
      break;
    }
    case 'share_users': {
      onlineUsersMap.value = dataObj;
      break;
    }
    case 'perm_expired': {
      const warningMsg = `${t('PermissionExpired')}: ${dataObj.detail}`;
      message.warning(warningMsg);
      warningIntervalId.value = window.setInterval(() => {
        message.warning(warningMsg);
      }, 1000 * 31);
      break;
    }
    case 'perm_valid': {
      if (warningIntervalId.value) {
        window.clearInterval(warningIntervalId.value);
        warningIntervalId.value = null;
      }
      message.success(t('PermissionValid'));
      break;
    }
    default:
      break;
  }
};

const onClientError = (status: any) => {
  loading.value = false;
  console.error('Guacamole client error:', status);
  const code = status.code;
  let msg = status.message || t('UnknownError');
  const currentLang = LanguageCode;
  msg = ErrorStatusCodes[code]
    ? t(ErrorStatusCodes[code])
    : t(ConvertGuacamoleError(status.message));
  console.log('Guacamole error message:', msg);
  switch (code) {
    case 1005:
      // 管理员终断会话，特殊处理
      if (currentLang === 'cn') {
        msg = status.message + ' ' + msg;
      } else {
        msg = msg + ' ' + status.message;
      }
      break;
    case 1003:
    case 1010:
      msg = msg.replace('{PLACEHOLDER}', status.message);
      break;
    case 1006:
      msg = msg + ': ' + status.message;
      break;
  }
  message.error(msg);
};

const onFileSystem = (obj: object, name: any) => {
  enableFilesystem.value = true;
  console.log('Guacamole file system initialized:', obj, name);
};

const onfile = (stream: number, mimetype: string, name: string) => {
  console.log('Guacamole file received:', stream, mimetype, name);
  // Handle file reception logic here
  // For example, you can display the file or process it as needed
  // This is a placeholder for actual file handling logic
};

const onclipboard = (stream: object, mimetype: string) => {
  let reader: any = null;
  // If the received data is text, read it as a simple string
  if (/^text\//.exec(mimetype)) {
    reader = new Guacamole.StringReader(stream);

    // Assemble received data into a single string
    let data = '';
    reader.ontext = (text: any) => {
      data += text;
    };

    // Set clipboard contents once stream is finished
    reader.onend = async () => {
      console.log('clipboard received from remote: ', data);
      if (navigator.clipboard) {
        remoteClipboardText.value = data;
        await navigator.clipboard.writeText(data);
      }
    };
  }
  // Otherwise read the clipboard data as a Blob
  else {
    reader = new Guacamole.BlobReader(stream, mimetype);
    reader.onprogress = (text: any) => {
      console.log('clipboard blob received from remote: ', text);
    };
    reader.onend = () => {
      remoteClipboardText.value = reader.getBlob();
      console.log('clipboard blob received from remote: ', remoteClipboardText);
    };
  }
};

const fileDrop = (event: any) => {
  event.stopPropagation();
  event.preventDefault();
  const files = event.dataTransfer.files;
  if (files.length === 0) {
    return;
  }
  console.log('Files dropped:', files);
  // Handle file drop logic here
  // For example, you can upload the files or process them as needed
  // This is a placeholder for actual file handling logic
};

const debouncedSendClipboardToRemote = useDebounceFn(async () => {
  const text = await readClipboardText();
  if (!text || !text.trim()) {
    return;
  }
  sendTextToRemote(text);
}, 300);

const connectGuacamole = async (connectString: string) => {
  const displayRef: any = document.getElementById('display');
  const tunnel = new Guacamole.WebSocketTunnel(wsPrefix.value);
  tunnel.receiveTimeout = 60 * 1000; // Set receive timeout to 60 seconds
  const client = new Guacamole.Client(tunnel);
  tunnel.onerror = (error: any) => {
    message.error(t('WebSocketError') + ` tunnel : ${error.message}`);
  };
  tunnel.onuuid = (uuid: string) => {
    tunnel.uuid = uuid;
    console.log('WebSocket UUID:', uuid);
  };
  client.onfilesystem = onFileSystem;
  client.onfile = onfile;

  client.onstatechange = clientStateChanged;
  client.onerror = onClientError;
  client.onclipboard = onclipboard;

  const oninstruction = tunnel.oninstruction;
  tunnel.oninstruction = (opcode: any, argv: any) => {
    if (oninstruction) {
      oninstruction(opcode, argv);
    }
    if (opcode === 'jms_event') {
      onJmsEvent(argv[0], argv[1]);
    }
  };

  displayRef.addEventListener(
    'dragenter',
    function (e: any) {
      e.stopPropagation();
      e.preventDefault();
    },
    false,
  );
  displayRef.addEventListener(
    'dragover',
    function (e: any) {
      e.stopPropagation();
      e.preventDefault();
    },
    false,
  );
  displayRef.addEventListener('drop', fileDrop, false);
  guacTunnel.value = tunnel;
  guacClient.value = client;
  const display = client.getDisplay();
  guacDisplay.value = display;
  const displayEl = display.getElement();
  displayEl.onclick = (e: any) => {
    e.preventDefault();
    return false;
  };

  display.onresize = () => {
    updateScale();
  };
  display.oncursor = (canvas: any, x: any, y: any) => {
    localCursor.value = true;
  };

  const mouse = new Guacamole.Mouse(displayRef);
  const sendScaledMouseState = (mouseState: any) => {
    const scaledState = new Guacamole.Mouse.State(
      mouseState.x / display.getScale(),
      mouseState.y / display.getScale(),
      mouseState.left,
      mouseState.middle,
      mouseState.right,
      mouseState.up,
      mouseState.down,
    );
    client.sendMouseState(scaledState);
  };

  const handleEmulatedMouseDown = (mouseState: any) => {
    // Emulate mouse down event
    if (client || display) {
      return;
    }
    lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '');
    // Send mouse state, show cursor if necessary
    display.showCursor(true);
    sendScaledMouseState(mouseState);
  };
  const handleEmulatedMouseState = (mouseState: any) => {
    // Emulate mouse move/up event
    if (client || display) {
      return;
    }
    lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '');
    // Send mouse state, hide cursor if necessary
    display.showCursor(true);
    sendScaledMouseState(mouseState);
  };

  mouse.onmousedown =
    mouse.onmouseup =
    mouse.onmousemove =
      (mouseState: any) => {
        // Send mouse state, hide cursor if necessary
        sendScaledMouseState(mouseState);
      };
  mouse.onmouseout = (mouseState: any) => {
    // Send mouse state, hide cursor if necessary
    display.showCursor(false);
  };

  const touchScreen = new Guacamole.Mouse.Touchscreen(displayEl);
  touchScreen.onmousedown = handleEmulatedMouseDown;
  touchScreen.onmousemove = touchScreen.onmouseup = handleEmulatedMouseState;
  const sink = new Guacamole.InputSink();
  const guacKeyboard = new Guacamole.Keyboard(sink.getElement());
  // guacKeyboard.listenTo(sink.getElement());
  guacKeyboard.onkeydown = (keysym: any) => {
    lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '');
    client.sendKeyEvent(1, keysym);
  };
  guacKeyboard.onkeyup = (keysym: any) => {
    lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '');
    client.sendKeyEvent(0, keysym);
  };

  displayRef.appendChild(display.getElement());
  document.body.appendChild(sink.getElement());

  window.addEventListener('focus', debouncedSendClipboardToRemote, false);

  const handleMouseEnter = () => {
    if (displayEl) displayEl.style.cursor = 'none';
    display.showCursor(true);
    sink.focus();
    document.body.focus();
  };
  const handleMouseLeave = () => {
    if (displayEl) displayEl.style.cursor = 'default';
  };
  displayEl.addEventListener('mouseenter', handleMouseEnter);
  displayEl.addEventListener('mouseleave', handleMouseLeave);
  client.connect(connectString);
};

onMounted(async () => {
  loading.value = true;
  const handLunaOpen = (message: any) => {
    console.log('Received Luna command:', message);
    drawShow.value = !drawShow.value;
  };
  lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.OPEN, handLunaOpen);
  const params = getCurrentConnectParams();
  wsPrefix.value = params.ws || '';
  apiPrefix.value = params.api || '';
  const token = params['data'].token || '';
  const connectString = await getConnectString(token);
  await connectGuacamole(connectString);
});

onUnmounted(() => {
  if (guacClient.value) {
    guacClient.value.disconnect();
  }
  lunaCommunicator.offLuna(LUNA_MESSAGE_TYPE.OPEN);
  lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.CLOSE, '');
  window.removeEventListener, 'focus';
});

const ClipBoardTextChange = (text: string) => {
  if (!text || !text.trim()) {
    return;
  }
  console.log('ClipBoardTextChange:', text);
  sendTextToRemote(text);
};

const sendTextToRemote = (text: string) => {
  const data = {
    type: 'text/plain',
    data: text,
  };
  if (!guacClient.value) {
    console.warn('Guacamole client is not initialized yet.');
    return;
  }
  let writer: any = null;
  const stream = guacClient.value.createClipboardStream(data.type);
  // Send data as a string if it is stored as a string
  if (typeof data.data === 'string') {
    writer = new Guacamole.StringWriter(stream);
    writer.sendText(data.data);
    writer.sendEnd();
  } else {
    // Write File/Blob asynchronously
    writer = new Guacamole.BlobWriter(stream);
    writer.oncomplete = function clipboardSent() {
      writer.sendEnd();
    };
    // Begin sending data
    writer.sendBlob(data.data);
  }
};
</script>

<template>
  <div class="w-full h-full justify-center flex flex-col">
    <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
      <n-spin :show="loading" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
      </n-spin>
    </div>
    <div id="display" v-show="!loading" class="w-screen h-screen"></div>
  </div>
  <n-drawer v-model:show="drawShow" :min-width="502" animated resizable>
    <n-drawer-content>
      <n-tabs default-value="settings" justify-content="space-evenly" type="line">
        <n-tab-pane name="settings" tab="设置">
          <ClipBoardText
            :disabled="!hasClipboardPermission"
            :remote-text="remoteClipboardText"
            @update:text="ClipBoardTextChange"
          />
        </n-tab-pane>
        <n-tab-pane name="file-manager" tab="文件管理">
          <FileManager />
        </n-tab-pane>
        <n-tab-pane name="share-collaboration" tab="分享会话"> 分享会话 </n-tab-pane>
      </n-tabs>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
#display {
  display: flex;
  justify-content: center;
  position: relative;
}
</style>
