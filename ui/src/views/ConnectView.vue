<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch, h } from 'vue';
import { set, useWindowSize } from '@vueuse/core';
import { useDebounceFn } from '@vueuse/core';
// @ts-ignore
// import Guacamole from 'guacamole-common-js';
import Guacamole from '@dushixiang/guacamole-common-js';
import type { UploadCustomRequestOptions, UploadFileInfo, UploadSettledFileInfo } from 'naive-ui';
import { NSpin, useMessage, NTabPane } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { getCurrentConnectParams, BaseAPIURL } from '@/utils/common';
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
  createAudioStream: (type: string) => any;
  sendKeyEvent: (pressed: number, keyCode: number, keyChar?: string) => void;
}

interface GuacamoleTunnel {
  uuid: string;
  connect: () => void;
  disconnect: () => void;
}

const FileType = {
  NORMAL: 'NORMAL',
  DIRECTORY: 'DIRECTORY',
};

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
const fileFsloading = ref(false);

const remoteClipboardText = ref<string>('');
const hasClipboardPermission = ref(false);
const enableFilesystem = ref(false);
const autoFit = ref<boolean>(true);
const debouncedResize = useDebounceFn(() => {
  updateScale();
  if (!autoFit.value) {
    return;
  }
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
    fitPercentage.value = Math.floor(newScale * 100);
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
      requestAudioStream(guacClient.value);
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

const requestAudioStream = (client: any) => {
  if (!guacClient.value || !guacClient.value.createClipboardStream) {
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
  recorder.onclose = requestAudioStream.bind(this, client); // 重新请求音频流
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

const currentGuacFsObject = ref<any>(null);
const currentGuacFsName = ref<any>(null);

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
const currentFolderObject = ref<any>(null);
const current_files = ref<any>({});
const currentFolder = ref<GuacamoleFile>();
const currentFolderFiles = ref<any>([]);

const handleFolderOpen = (row: any) => {
  console.log('Opening folder:', row);
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
      console.log('Current folder files:', currentFolderFiles.value);
    })
    .catch((error: any) => {
      console.error('Error refreshing folder:', error);
      message.error(t('FileSystemError') + ': ' + error.message);
    })
    .finally(() => {
      fileFsloading.value = false;
    });
};

const handleDownloadFile = (fileItem: GuacamoleFile) => {
  console.log('Downloading file:', fileItem);
  if (!fileItem || !fileItem.streamName) {
    console.warn('Cannot download file, file is not valid:', fileItem);
    return;
  }
  const path = fileItem.streamName;
  const downloadStream = (stream: any, mimetype: any) => {
    clientFileReceived(stream, mimetype, fileItem.name);
  };
  currentGuacFsObject.value.requestInputStream(path, downloadStream);
};

interface UploadItem {
  uploadOptions: UploadCustomRequestOptions;
  folder: GuacamoleFile;
}

const uploadingFiles = ref<Array<UploadItem>>([]);
const isUploading = ref(false);

const displayUploadingFiles = ref<Array<UploadSettledFileInfo>>([]);

const handleUploadFile = (options: UploadCustomRequestOptions, folder: any) => {
  if (!options.file || !currentGuacFsObject.value) {
    console.warn('Cannot upload file, file or Guacamole object is not valid:', options.file);
    return;
  }
  const item = {
    uploadOptions: options,
    folder: folder || currentFolder.value,
  };
  displayUploadingFiles.value.push(options.file);
  uploadingFiles.value.push(item);
  if (isUploading.value) {
    console.warn('Already uploading files, skipping new upload:', options.file.name);
    return;
  }
  isUploading.value = true;

  processUploadQueue().then(() => {
    handleFolderOpen(currentFolder.value);
  });
};

const handleRemoveFile = (file: any) => {
  if (file.status === 'uploading') {
    message.warning(t('FileUploadingWarning'));
    return;
  }
  const newDisplayFiles = displayUploadingFiles.value.filter((f) => {
    return f.name !== file.name;
  });
  nextTick(() => {
    displayUploadingFiles.value = newDisplayFiles;
  });
};

const processUploadQueue = async () => {
  while (isUploading.value && uploadingFiles.value.length > 0) {
    const UploadItem = uploadingFiles.value.shift();
    if (!UploadItem || !UploadItem.uploadOptions) {
      continue;
    }
    const { uploadOptions, folder } = UploadItem;

    try {
      uploadOptions.file.status = 'uploading';
      await uploadFile(uploadOptions, folder);
      uploadOptions.file.status = 'finished';
      setTimeout(() => {
        handleRemoveFile(uploadOptions.file);
      }, 1000 * 5); // 延迟5秒后移除上传文件
    } catch (error) {
      console.error('Error processing upload queue:', error);
      message.error(t('FileUploadError') + ': ' + error);
      uploadOptions.file.status = 'error';
    }
  }
  isUploading.value = false;
};

const fakeProcessInterval = ref<number | null>(null);

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
  if (!guacClient.value || !guacClient.value.createClipboardStream) {
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

import {
  NFlex,
  NButton,
  NInput,
  NText,
  NScrollbar,
  NDataTable,
  NCard,
  NModal,
  NUpload,
  NUploadTrigger,
  NDrawer,
  NDrawerContent,
  NPopover,
  NIcon,
  NProgress,
  useNotification,
} from 'naive-ui';

import type { NotificationOptions, NotificationReactive } from 'naive-ui';
const notification = useNotification();
const nRef = ref<NotificationReactive | null>(null);

const uploadGuacamoleFile = (
  file: any,
  object: any,
  streamName: any,
  progressCallback: CallableFunction,
): Promise<void> => {
  const clinet = guacClient.value;
  const tunnel = guacTunnel.value;
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

const sanitizeFilename = (filename: string) => {
  return filename.replace(/[\\\/]+/g, '_');
};
const clientFileReceived = (stream: any, mimetype: any, filename: any) => {
  // Build download URL
  const uuid = guacTunnel.value?.uuid || '';
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

const onFileSystem = (obj: any, name: any) => {
  if (!obj || !Guacamole.Object) {
    console.warn('Guacamole file system object or name is not provided.');
    return;
  }
  console.log('Guacamole file system object received:', obj, name);

  enableFilesystem.value = true;
  currentGuacFsObject.value = obj;
  currentFolderObject.value = obj;
  currentGuacFsName.value = name;
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

const onfile = (stream: number, mimetype: string, name: string) => {
  clientFileReceived(stream, mimetype, name);
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

const showOsk = ref<boolean>(false);

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

  const mouse = new Guacamole.Mouse(display.getElement());
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
    nextTick(() => {
      drawShow.value = !drawShow.value;
    });
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
document.addEventListener(
  'contextmenu',
  (e: MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
  },
  false,
);
import Osk from '@/components/Osk.vue';
import KeyboardOption from '@/components/KeyboardOption.vue';
import OtherOption from '@/components/OtherOption.vue';
const keyboardLayout = ref<string>('en-us-qwerty');

const handleScreenKeyboard = (name: string, keysym: any) => {
  console.log('Screen keyboard change:', name, keysym);
  if (!guacClient.value) {
    console.warn('Guacamole client is not initialized yet.');
    return;
  }
  switch (name) {
    case 'keydown':
      guacClient.value.sendKeyEvent(1, keysym);
      break;
    case 'keyup':
      guacClient.value.sendKeyEvent(0, keysym);
      break;
    default:
      console.warn('Unknown screen keyboard event:', name);
  }
};

const fitPercentage = ref<number>(100);

watch(
  [autoFit],
  ([newAutoFit]) => {
    if (newAutoFit) {
      debouncedResize();
    }
  },
  { immediate: true },
);

const handleCombineKeys = (keys: string[]) => {
  if (!guacClient.value) {
    console.warn('Guacamole client is not initialized yet.');
    return;
  }
  const client = guacClient.value;
  keys.forEach((keysym: any) => {
    client.sendKeyEvent(1, keysym);
  });
  setTimeout(() => {
    keys.forEach((keysym: any) => {
      client.sendKeyEvent(0, keysym);
    });
  }, 100);
};

const scaleGuacDisplay = (value: number) => {
  if (!guacDisplay.value) {
    console.warn('Guacamole display is not initialized yet.');
    return;
  }
  if (value <= 0) {
    console.warn('Invalid scale value:', scale);
    return;
  }
  console.log('Scaling Guacamole display to:', value);
  const newScale = value / 100; // 限制缩放范围在0.1到5之间

  guacDisplay.value.scale(newScale);
  scale.value = newScale;
  fitPercentage.value = Math.floor(newScale * 100);
};
</script>

<template>
  <div class="w-full h-full justify-center flex flex-col">
    <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
      <n-spin :show="loading" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
      </n-spin>
    </div>
    <div id="display" v-show="!loading" class="w-screen h-screen"></div>
    <Osk v-if="showOsk" :keyboard="keyboardLayout" @keyboard-change="handleScreenKeyboard" />
  </div>

  <n-drawer v-model:show="drawShow" :min-width="502" :default-width="502" resizable>
    <n-drawer-content>
      <n-tabs default-value="settings" justify-content="space-evenly" type="line">
        <n-tab-pane name="settings" tab="设置">
          <ClipBoardText
            :disabled="!hasClipboardPermission"
            :remote-text="remoteClipboardText"
            @update:text="ClipBoardTextChange"
          />
          <br />
          <KeyboardOption v-model:opened="showOsk" v-model:keyboard="keyboardLayout" />
          <br />
          <OtherOption
            v-model:auto-fit="autoFit"
            :fit-percentage="fitPercentage"
            @combine-keys="handleCombineKeys"
            @update-scale="scaleGuacDisplay"
            :is-remote-app="false"
          />
        </n-tab-pane>
        <n-tab-pane name="file-manager" tab="文件管理">
          <FileManager
            :loading="fileFsloading"
            :files="currentFolderFiles"
            :name="currentGuacFsName"
            :folder="currentFolder"
            :display-uploading-files="displayUploadingFiles"
            @open-folder="handleFolderOpen"
            @download-file="handleDownloadFile"
            @upload-file="handleUploadFile"
            @remove-upload-file="handleRemoveFile"
          />
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
