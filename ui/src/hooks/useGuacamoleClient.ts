import { ref } from 'vue';
// @ts-ignore
import Guacamole from 'guacamole-common-js';

import { useDebounceFn } from '@vueuse/core';
import { BaseAPIURL } from '@/utils/common';
import { readClipboardText } from '@/utils/clipboard';
import { NSpin, useMessage, NTabPane } from 'naive-ui';
import type { UploadCustomRequestOptions, UploadFileInfo, UploadSettledFileInfo } from 'naive-ui';
const supportImages: any[] = [];
const pendingTests: any[] = [];
const testImages: any = {
  /**
   * Test JPEG image, encoded as base64.
   */
  'image/jpeg':
    '/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoH' +
    'BwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRT/2wBDAQME' +
    'BAUEBQkFBQkUDQsNFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQU' +
    'FBQUFBQUFBQUFBQUFBT/wAARCAABAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAA' +
    'AAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUAQEAAAAAAAAAAAAAAA' +
    'AAAAAA/8QAFBEBAAAAAAAAAAAAAAAAAAAAAP/aAAwDAQACEQMRAD8AVMH/2Q==',

  /**
   * Test PNG image, encoded as base64.
   */
  'image/png':
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEX///+nxBvI' +
    'AAAACklEQVQI12NgAAAAAgAB4iG8MwAAAABJRU5ErkJggg==',

  /**
   * Test WebP image, encoded as base64.
   */
  'image/webp': 'UklGRhoAAABXRUJQVlA4TA0AAAAvAAAAEAcQERGIiP4HAA==',
}; // 测试单个图片格式
async function testImageFormat(mimeType: string, base64Data: any): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    const image = new Image();

    image.onload = () => {
      // Image format is supported if successfully decoded with correct dimensions
      const isSupported = image.width === 1 && image.height === 1;
      resolve(isSupported);
    };

    image.onerror = () => {
      console.debug(`Format ${mimeType} not supported`);
      resolve(false);
    };

    // Set source to trigger loading
    image.src = `data:${mimeType};base64,${base64Data}`;
  });
}

// Use Object.entries for better iteration over key-value pairs
Object.entries(testImages).forEach(([mimeType, base64Data]) => {
  const imageTest = new Promise<void>((resolve) => {
    const image = new Image();

    // Set up handlers before setting src to avoid race conditions
    image.onload = () => {
      // Image format is supported if successfully decoded with correct dimensions
      if (image.width === 1 && image.height === 1) {
        supportImages.push(mimeType);
      }
      resolve();
    };

    // Handle errors separately for better debugging
    image.onerror = () => {
      console.debug(`Format ${mimeType} not supported`);
      resolve(); // Still resolve to continue testing other formats
    };

    // Set source to trigger loading
    image.src = `data:${mimeType};base64,${base64Data}`;
  });

  pendingTests.push(imageTest);
});
const FileType = {
  NORMAL: 'NORMAL',
  DIRECTORY: 'DIRECTORY',
};

export async function getSupportedImages(): Promise<string[]> {
  // 清空之前的结果
  supportImages.length = 0;

  // 并行测试所有图片格式
  const testPromises = Object.entries(testImages).map(async ([mimeType, base64Data]) => {
    const isSupported = await testImageFormat(mimeType, base64Data);
    if (isSupported) {
      supportImages.push(mimeType);
    }
    return { mimeType, isSupported };
  });

  // 等待所有测试完成
  await Promise.all(testPromises);

  return [...supportImages]; // 返回副本
}
export async function getSupportedGuacVideos(): Promise<string[]> {
  return Guacamole.VideoPlayer.getSupportedTypes();
}
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';

import { LanguageCode } from '@/locales';

import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { ErrorStatusCodes, ConvertGuacamoleError } from '@/utils/status';
import { nextTick } from 'vue';
export async function getSupportedGuacAudios(): Promise<string[]> {
  return Guacamole.AudioPlayer.getSupportedTypes();
}

export async function getSupportedGuacMimeTypes(): Promise<string> {
  const supportImages = await getSupportedImages();
  const supportVideos = await getSupportedGuacVideos();
  const supportAudios = await getSupportedGuacAudios();
  let connectString = '';
  supportImages.forEach((mimeType) => {
    connectString += '&GUAC_IMAGE=' + encodeURIComponent(mimeType);
  });
  supportVideos.forEach((mimeType) => {
    connectString += '&GUAC_VIDEO=' + encodeURIComponent(mimeType);
  });
  supportAudios.forEach((mimeType) => {
    connectString += '&GUAC_AUDIO=' + encodeURIComponent(mimeType);
  });
  return connectString;
}

export async function getSupportedMimeTypes(): Promise<Record<string, string[]>> {
  const supportImages = await getSupportedImages();
  const supportVideos = await getSupportedGuacVideos();
  const supportAudios = await getSupportedGuacAudios();
  return {
    GUAC_IMAGE: supportImages,
    GUAC_VIDEO: supportVideos,
    GUAC_AUDIO: supportAudios,
  };
}

const sanitizeFilename = (filename: string) => {
  return filename.replace(/[\\\/]+/g, '_');
};

interface GuacamoleFile {
  mimetype?: any;
  streamName?: any;
  type: 'DIRECTORY' | 'FILE';
  name: string;
  parent?: GuacamoleFile | null;
  is_dir?: boolean;
}

export function useGuacamoleClient(t: any) {
  const pixelDensity = window.devicePixelRatio || 1;
  const guaClient = ref<Guacamole.Client | null>(null);
  const guaTunnel = ref<Guacamole.Tunnel | null>(null);
  const guaDisplay = ref<Guacamole.Display | null>(null);
  const fsObject = ref<Guacamole.File | null>(null);
  const driverName = ref<string>('');
  const connectStatus = ref('Connecting');
  const sessionObject = ref<any>({});
  const action_permission = ref<any>({});
  const enableShare = ref(false);
  const hasClipboardPermission = ref(false);
  const currentUser = ref<any>({});
  const shareId = ref<string | null>(null);
  const onlineUsersMap = ref<Record<string, any>>({});
  const warningIntervalId = ref<number | null>(null);
  const loading = ref(true);
  const scale = ref(1);
  const currentWidth = ref(window.innerWidth);
  const currentHeight = ref(window.innerHeight);
  const fakeProcessInterval = ref<number | null>(null);
  const message = useMessage();
  const sink = new Guacamole.InputSink();
  const keyboard = new Guacamole.Keyboard();
  function connectToGuacamole(
    wsUrl: string,
    connectParams: Record<string, any>,
    width: any,
    height: any,
    supportFs: boolean = false,
  ) {
    currentWidth.value = width || window.innerWidth;
    currentHeight.value = height || window.innerHeight;

    const tunnel = new Guacamole.WebSocketTunnel(wsUrl);
    tunnel.receiveTimeout = 60 * 1000; // Set receive timeout to 60 secondsa
    const client = new Guacamole.Client(tunnel);

    tunnel.onerror = (error: any) => {
      const code = error.code || 0;
      const messageText = error.message || t('WebSocketError');
      message.error(t('WebSocketError'));
    };
    tunnel.onuuid = (uuid: string) => {
      tunnel.uuid = uuid;
      console.log('WebSocket UUID:', uuid);
    };

    const oninstruction = tunnel.oninstruction;
    tunnel.oninstruction = (opcode: any, argv: any) => {
      if (oninstruction) {
        oninstruction(opcode, argv);
      }
      if (opcode === 'jms_event') {
        onJmsEvent(argv[0], argv[1]);
      }
    };
    if (supportFs) {
      client.onfilesystem = onFileSystem;
      client.onfile = clientFileReceived;
    }

    client.onstatechange = clientStateChanged;
    client.onerror = onClientError;
    client.onclipboard = onclipboard;
    const display = client.getDisplay();
    display.onresize = (resizeEvent: any) => {
      console.log('Guacamole display resized:', resizeEvent);
      updateScale();
    };
    display.showCursor(false);
    guaDisplay.value = display;
    guaClient.value = client;
    guaTunnel.value = tunnel;
    const queryParams = new URLSearchParams(connectParams);
    if (width) {
      queryParams.append('GUAC_WIDTH', width.toString());
    }
    if (height) {
      queryParams.append('GUAC_HEIGHT', height.toString());
    }
    const optimalDpi = pixelDensity * 96;
    queryParams.append('GUAC_DPI', optimalDpi.toString());
    getSupportedMimeTypes()
      .then((mimeTypes: Record<string, string[]>) => {
        // add supported mime types to query params
        Object.entries(mimeTypes).forEach(([key, values]) => {
          values.forEach((value) => {
            queryParams.append(key, value);
          });
        });
      })
      .finally(() => {
        // Connect the client after adding all supported mime types
        console.log('Connecting to Guacamole with params:', queryParams.toString());
        client.connect(queryParams.toString());
      });
  }

  const disconnectGuaclient = () => {
    if (guaClient.value) {
      guaClient.value.disconnect();
    }
  };

  const registerMouseAndKeyboardHanlder = () => {
    const client: Guacamole.Client = guaClient.value;
    if (!client || !client.getDisplay) {
      return console.warn(
        'Guacamole client is not initialized or does not support mouse and keyboard events',
      );
    }
    window.addEventListener('focus', debouncedSendClipboardToRemote, false);
    registerMouse(client);

    registerTouchScreen(client);

    registerKeyboard(client);
    const display = client.getDisplay();
    const displayEl = display.getElement();

    const handleMouseEnter = () => {
      if (displayEl) displayEl.style.cursor = 'none';
      display.showCursor(true);
      document.body.focus();
      nextTick(() => {
        sink.focus();
      });
    };
    const handleMouseLeave = () => {
      if (displayEl) displayEl.style.cursor = 'default';
      nextTick(() => {
        keyboard.reset();
      });
    };
    displayEl.addEventListener('mouseenter', handleMouseEnter);
    displayEl.addEventListener('mouseleave', handleMouseLeave);
  };

  const resizeGuaScale = useDebounceFn((width: number, height: number) => {
    currentWidth.value = width || window.innerWidth;
    currentHeight.value = height || window.innerHeight;
    updateScale();
  }, 300);

  const sendGuaSize = (width: number, height: number) => {
    if (guaClient.value && guaDisplay.value) {
      console.log('Sending resize to Guacamole client:', width, height);
      guaClient.value.sendSize(width, height);
    }
  };
  const updateScale = () => {
    if (!guaDisplay.value || !guaClient.value) {
      console.warn('Guacamole display is not initialized yet.');
      return;
    }
    const w = guaDisplay.value.getWidth();
    const h = guaDisplay.value.getHeight();

    if (h === 0 || w === 0) {
      return 1;
    }
    const newScale = Math.min(currentWidth.value / w, currentHeight.value / h);
    if (newScale !== scale.value) {
      console.log(`Guacamole display scaled from ${scale.value} to ${newScale}`);
      scale.value = newScale;
      guaDisplay.value.scale(newScale);
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
        console.log('Online users updated:', onlineUsersMap.value);
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
  const debouncedSendClipboardToRemote = useDebounceFn(async () => {
    const text = await readClipboardText();
    if (!text || !text.trim()) {
      return;
    }
    sendTextToRemote(text);
  }, 300);
  const sendTextToRemote = (text: string) => {
    const data = {
      type: 'text/plain',
      data: text,
    };
    if (!guaClient.value) {
      console.warn('Guacamole client is not initialized yet.');
      return;
    }
    let writer: any = null;
    const stream = guaClient.value.createClipboardStream(data.type);
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

  const registerKeyboard = (client: Guacamole.Client) => {
    if (!client || !client.getDisplay) {
      console.warn('Guacamole client is not initialized or does not support keyboard events');
      return;
    }
    const display = client.getDisplay();
    if (!display) {
      console.warn('Guacamole display is not available');
      return;
    }
    keyboard.listenTo(sink.getElement());

    keyboard.onkeydown = (keysym: any) => {
      client.sendKeyEvent(1, keysym);
      lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '');
    };
    keyboard.onkeyup = (keysym: any) => {
      client.sendKeyEvent(0, keysym);
    };
    display.getElement().appendChild(sink.getElement());
  };

  const registerTouchScreen = (client: Guacamole.Client) => {
    if (!client || !client.getDisplay) {
      console.warn('Guacamole client is not initialized or does not support screen events');
      return;
    }
    const display = client.getDisplay();
    if (!display) {
      console.warn('Guacamole display is not available');
      return;
    }
    const touchScreen = new Guacamole.Mouse.Touchscreen(display.getElement());
    const handleEmulatedMouseDown = (mouseState: any) => {
      // Emulate mouse down event
      if (client || display) {
        return;
      }
      lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '');
      // Send mouse state, show cursor if necessary
      display.showCursor(true);
      sendScaledMouseState(client, mouseState);
    };

    const handleEmulatedMouseState = (mouseState: any) => {
      // Emulate mouse move/up event
      if (client || display) {
        return;
      }
      lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '');
      // Send mouse state, hide cursor if necessary
      display.showCursor(true);
      sendScaledMouseState(client, mouseState);
    };
    touchScreen.onmousedown = handleEmulatedMouseDown;
    touchScreen.onmousemove = touchScreen.onmouseup = handleEmulatedMouseState;
  };

  const sendScaledMouseState = (client: any, mouseState: any) => {
    const display = client.getDisplay();
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

  const sendKeyEvent = (released: number, keysym: number) => {
    if (!guaClient.value) {
      console.warn('Guacamole client is not initialized yet.');
      return;
    }

    guaClient.value.sendKeyEvent(released, keysym);
  };

  const registerMouse = (client: Guacamole.Client) => {
    if (!client || !client.getDisplay) {
      console.warn('Guacamole client is not initialized or does not support mouse events');
      return;
    }
    const display = client.getDisplay();
    if (!display) {
      console.warn('Guacamole display is not available');
      return;
    }
    const sendMouseState = (mouseState: any) => {
      sendScaledMouseState(client, mouseState);
    };
    const mouse = new Guacamole.Mouse(display.getElement());
    mouse.onmousedown =
      mouse.onmouseup =
      mouse.onmousemove =
        (mouseState: any) => {
          // Send mouse state, hide cursor if necessary
          sendMouseState(mouseState);
        };
    mouse.onmouseout = (mouseState: any) => {
      // Send mouse state, hide cursor if necessary
      display.showCursor(false);
    };
  };

  const onClientError = (status: any) => {
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
        requestAudioStream(guaClient.value);
        break;
      case 4:
        connectStatus.value = 'Disconnecting';
        break;
      case 5:
        connectStatus.value = 'Disconnected';
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.CLOSE, '');
        guaDisplay.value?.getElement()?.remove();
        break;
    }
  };
  const onFileSystem = (obj: any, name: any) => {
    if (!obj || !Guacamole.Object) {
      console.warn('Guacamole file system object or name is not provided.');
      return;
    }
    console.log('Guacamole file system object received:', obj, name);

    enableFilesystem.value = true;
    fsObject.value = obj;
    currentFolderObject.value = obj;
    driverName.value = name;
    currentGuacFsObject.value = obj;
    const defaultFolder: GuacamoleFile = {
      mimetype: Guacamole.Object.STREAM_INDEX_MIMETYPE,
      streamName: Guacamole.Object.ROOT_STREAM,
      type: 'DIRECTORY',
      is_dir: true,
      name: name,
      parent: null,
    };
    currentFolder.value = defaultFolder;
    handleFolderOpen(defaultFolder);
  };
  const requestAudioStream = (client: any) => {
    if (!client || !client.createClipboardStream) {
      console.warn('Guacamole client is not initialized or does not support audio stream');
      return;
    }
    const AUDIO_INPUT_MIMETYPE = 'audio/L16;rate=44100,channels=2';
    const audioStream = client.createAudioStream(AUDIO_INPUT_MIMETYPE);
    const recorder = Guacamole.AudioRecorder.getInstance(audioStream, AUDIO_INPUT_MIMETYPE);
    if (!recorder) {
      audioStream.sendEnd();
      return;
    }
    recorder.onclose = () => {
      console.log('Audio stream closed');
      requestAudioStream(client); // 重新请求音频流
    }; // 重新请求音频流
  };
  const currentFolderFiles = ref<any>([]);
  const current_files = ref<any>({});
  const currentFolder = ref<GuacamoleFile | null>(null);
  const currentFolderObject = ref<Guacamole.File | null>(null);
  const fileFsloading = ref(false);
  const enableFilesystem = ref(false);
  const currentGuacFsObject = ref<Guacamole.File | null>(null);
  const handleFolderOpen = (row: any) => {
    if (!row || !row.is_dir) {
      console.warn('Cannot open folder, row is not a directory:', row);
      return;
    }
    currentFolder.value = row;
    currentFolderObject.value = row;
    fileFsloading.value = true;
    RefreshFileSystem(currentGuacFsObject.value, row)
      .then((files: any) => {
        current_files.value = files;
        currentFolderFiles.value = [] as GuacamoleFile[];
        for (const fileName in files) {
          currentFolderFiles.value.push({
            name: fileName,
            is_dir: files[fileName].type === 'DIRECTORY',
            mimetype: files[fileName].mimetype,
            streamName: files[fileName].streamName,
            parent: row,
          });
        }
        currentFolderFiles.value.sort((a: GuacamoleFile, b: GuacamoleFile) => {
          if (a.is_dir && !b.is_dir) {
            return -1; // Directories first
          } else if (!a.is_dir && b.is_dir) {
            return 1; // Files after directories
          }
          return a.name.localeCompare(b.name); // Sort alphabetically
        });
      })
      .catch((error: any) => {
        console.error('Error refreshing folder:', error);
        message.error(t('FileSystemError') + ': ' + error.message);
      })
      .finally(() => {
        fileFsloading.value = false;
      });
  };
  const clientFileReceived = (stream: any, mimetype: any, filename: any) => {
    // Build download URL
    const uuid = guaTunnel.value?.uuid || '';
    const url =
      BaseAPIURL +
      '/tunnels/' +
      encodeURIComponent(uuid) +
      '/streams/' +
      encodeURIComponent(stream.index) +
      '/' +
      encodeURIComponent(sanitizeFilename(filename));

    // Create temporary hidden iframe to facilitate download
    const iframe = document.createElement('iframe');
    iframe.style.position = 'fixed';
    iframe.style.border = 'none';
    iframe.style.width = '1px';
    iframe.style.height = '1px';
    iframe.style.left = '-1px';
    iframe.style.top = '-1px';

    // The iframe MUST be part of the DOM for the download to occur
    document.body.appendChild(iframe);

    // Automatically remove iframe from DOM when download completes, if
    // browser supports tracking of iframe downloads via the "load" event
    iframe.onload = function downloadComplete() {
      document.body.removeChild(iframe);
    };

    // Acknowledge (and ignore) any received blobs
    stream.onblob = function acknowledgeData() {
      stream.sendAck('OK', Guacamole.Status.Code.SUCCESS);
    };

    // Automatically remove iframe from DOM a few seconds after the stream
    // ends, in the browser does NOT fire the "load" event for downloads
    stream.onend = function downloadComplete() {
      window.setTimeout(function cleanupIframe() {
        if (iframe.parentElement) {
          document.body.removeChild(iframe);
        }
      }, 500);
    };
    // Begin download
    iframe.src = url;
  };

  const uploadGuacamoleFile = (
    file: any,
    object: any,
    streamName: any,
    progressCallback: CallableFunction,
  ): Promise<void> => {
    const clinet = guaClient.value;
    const tunnel = guaTunnel.value;
    if (!clinet || !tunnel) {
      return Promise.reject(new Error('Guacamole client or tunnel is not initialized'));
    }
    const uuid = tunnel.uuid;
    const stream = object.createOutputStream(file.type, streamName);
    return new Promise<void>((resolve, reject) => {
      stream.onack = function beginUpload(status: any) {
        // Notify of any errors from the Guacamole server
        if (status.isError()) {
          reject(status);
          return;
        }
        const uploadToStream = function uploadStream(
          tunnelId: any,
          stream: any,
          file: any,
          progressCallback: CallableFunction,
        ) {
          // Build upload URL
          const url =
            BaseAPIURL +
            '/tunnels/' +
            encodeURIComponent(tunnelId) +
            '/streams/' +
            encodeURIComponent(stream.index) +
            '/' +
            encodeURIComponent(sanitizeFilename(file.name));
          const xhr = new XMLHttpRequest();
          xhr.withCredentials = true;
          // Invoke provided callback if upload tracking is supported
          if (progressCallback && xhr.upload) {
            xhr.upload.addEventListener('progress', function updateProgress(e) {
              progressCallback(e);
            });
          }
          // Resolve/reject promise once upload has stopped
          xhr.onreadystatechange = () => {
            // Ignore state changes prior to completion
            if (xhr.readyState !== 4) {
              return;
            }
            // Resolve if HTTP status code indicates success
            if (xhr.status >= 200 && xhr.status < 300) {
              resolve();
            } else if (xhr.getResponseHeader('Content-Type') === 'application/json') {
              try {
                const error = JSON.parse(xhr.responseText);
                reject({ status: xhr.status, message: error.message });
              } catch (e) {
                reject({ status: xhr.status, message: 'Failed to parse error response' });
              }
            } else if (xhr.status >= 400 && xhr.status < 500) {
              reject({ status: xhr.status, message: xhr.responseText });
            } else {
              reject(xhr.status);
            }
          };
          // Perform upload
          xhr.open('POST', url, true);
          const fd = new FormData();
          fd.append('file', file);
          xhr.send(fd);
        };
        // Begin upload
        uploadToStream(uuid, stream, file, progressCallback);
        // Ignore all further acks
        stream.onack = null;
      };
    });
  };

  const uploadFile = async (options: UploadCustomRequestOptions, folder: GuacamoleFile) => {
    // 参数验证
    if (!options || !options.file) {
      const error = new Error('Upload options or file is missing');
      console.error('Upload failed:', error.message);
      message.error(t('FileUploadError') + ': ' + error.message);
      throw error;
    }

    if (!folder || !folder.streamName) {
      const error = new Error('Target folder is invalid or missing stream name');
      console.error('Upload failed:', error.message);
      message.error(t('FileUploadError') + ': ' + error.message);
      throw error;
    }

    // Guacamole 客户端验证
    if (!guaClient.value || !guaClient.value.createClipboardStream) {
      const error = new Error('Guacamole client is not initialized');
      console.error('Upload failed:', error.message);
      message.error(t('GuacamoleClientNotInitialized'));
      throw error;
    }

    // 文件系统对象验证
    if (!currentGuacFsObject.value) {
      const error = new Error('Guacamole file system object is not available');
      console.error('Upload failed:', error.message);
      message.error(t('FileSystemError') + ': ' + error.message);
      throw error;
    }

    const file = options.file;
    const { onFinish, onError } = options;

    // 文件名验证
    if (!file.name || file.name.trim() === '') {
      const error = new Error('File name is empty or invalid');
      console.error('Upload failed:', error.message);
      message.error(t('FileUploadError') + ': ' + error.message);
      throw error;
    }

    const streamName = folder.streamName + '/' + sanitizeFilename(file.name).replace(/[:]+/g, '_');
    console.log('Uploading file:', file.name, 'to stream:', streamName);

    if (fakeProcessInterval.value) {
      clearInterval(fakeProcessInterval.value);
      fakeProcessInterval.value = null;
    }

    fakeProcessInterval.value = setInterval(() => {
      if (file.percentage && file.percentage < 97) {
        file.percentage += 1;
      } else {
        if (fakeProcessInterval.value) {
          clearInterval(fakeProcessInterval.value);
          fakeProcessInterval.value = null;
        }
      }
    }, 1000 * 2);

    const progressCallback = (e: any) => {
      console.log('Upload progress:', e.loaded, '/', e.total, 'for file:', file.name);
      options.file.percentage = (e.loaded / e.total) * 100 - 40;
    };

    try {
      await uploadGuacamoleFile(file.file, currentGuacFsObject.value, streamName, progressCallback);
      onFinish();
      file.percentage = 100;
    } catch (error) {
      console.error('Error uploading file:', error);
      onError();
      throw error; // 重新抛出异常，让调用者知道上传失败
    } finally {
      // 确保无论成功还是失败都清理 interval
      if (fakeProcessInterval.value) {
        clearInterval(fakeProcessInterval.value);
        fakeProcessInterval.value = null;
      }
    }
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
          await navigator.clipboard.writeText(data);
        }
      };
    }
    // Otherwise read the clipboard data as a Blob
    else {
      reader = new Guacamole.BlobReader(stream, mimetype);
      reader.onprogress = (text: any) => {
        console.log('clipboard blob text received from remote: ', text);
      };
      reader.onend = () => {
        const blob = reader.getBlob();
        console.log('clipboard blob received from remote: ', blob);
        navigator.clipboard.write(blob);
      };
    }
  };
  interface GuacamoleFile {
    mimetype?: any;
    streamName?: any;
    type: 'DIRECTORY' | 'FILE';
    name: string;
    parent?: GuacamoleFile | null;
    is_dir?: boolean;
  }
  const RefreshFileSystem = (
    guacFsObject: any,
    file: GuacamoleFile,
  ): Promise<Record<string, GuacamoleFile>> => {
    if (!guacFsObject || !guacFsObject.requestInputStream || !file) {
      return Promise.reject(new Error('Guacamole guacFsObject is not initialized'));
    }
    return new Promise<Record<string, GuacamoleFile>>(function (resolve, reject) {
      // Do not attempt to refresh the contents of directories
      if (file.mimetype !== Guacamole.Object.STREAM_INDEX_MIMETYPE) {
        reject('Cannot refresh contents of file: ' + file.name);
        return;
      }
      // Request contents of given file
      guacFsObject.requestInputStream(
        file.streamName,
        function handleStream(stream: any, mimetype: any) {
          // Ignore stream if mimetype is wrong
          if (mimetype !== Guacamole.Object.STREAM_INDEX_MIMETYPE) {
            stream.sendAck('Unexpected mimetype', Guacamole.Status.Code.UNSUPPORTED);
            reject('Unexpected mimetype' + ': ' + mimetype + ' for file: ' + file.name);
            return;
          }

          // Signal server that data is ready to be received
          stream.sendAck('Ready', Guacamole.Status.Code.SUCCESS);

          // Read stream as JSON
          var reader = new Guacamole.JSONReader(stream);

          // Acknowledge received JSON blobs
          reader.onprogress = function onprogress() {
            stream.sendAck('Received', Guacamole.Status.Code.SUCCESS);
          };

          // Reset contents of directory
          reader.onend = function jsonReady() {
            // Empty contents
            const files: any = {};

            // Determine the expected filename prefix of each stream
            var expectedPrefix = file.streamName;
            if (expectedPrefix.charAt(expectedPrefix.length - 1) !== '/') {
              expectedPrefix += '/';
            }

            // For each received stream name
            var mimetypes = reader.getJSON();
            for (var name in mimetypes) {
              // Assert prefix is correct
              if (name.substring(0, expectedPrefix.length) !== expectedPrefix) {
                continue;
              }

              // Extract filename from stream name
              var filename = name.substring(expectedPrefix.length);
              // Deduce type from mimetype
              var type = FileType.NORMAL;
              if (mimetypes[name] === Guacamole.Object.STREAM_INDEX_MIMETYPE) {
                type = FileType.DIRECTORY;
              }

              // Add file entry
              files[filename] = {
                mimetype: mimetypes[name],
                streamName: name,
                type: type,
                parent: file,
                name: filename,
              };
            }
            resolve(files);
          };
          reader.onerror = function jsonError(error: any) {
            reject('Error reading JSON from Guacamole stream: ');
          };
        },
      );
    });
  };

  return {
    guaClient,
    guaTunnel,
    guaDisplay,
    connectToGuacamole,
    connectStatus,
    sessionObject,
    action_permission,
    enableShare,
    hasClipboardPermission,
    currentUser,
    shareId,
    onlineUsersMap,
    warningIntervalId,
    sanitizeFilename,
    getSupportedGuacMimeTypes,
    onJmsEvent,
    sendTextToRemote,
    debouncedSendClipboardToRemote,
    registerMouseAndKeyboardHanlder,
    registerTouchScreen,
    registerMouse,
    onClientError,
    clientStateChanged,
    onFileSystem,
    requestAudioStream,
    handleFolderOpen,
    clientFileReceived,
    onclipboard,
    RefreshFileSystem,
    loading,
    resizeGuaScale,
    sendKeyEvent,
    disconnectGuaclient,
    uploadGuacamoleFile,
    uploadFile,
    sendGuaSize,
    scale,
    driverName,
    currentFolder,
    currentFolderFiles,
    fileFsloading,
  };
}
