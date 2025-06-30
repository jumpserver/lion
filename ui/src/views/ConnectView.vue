<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue';
import { useWindowSize } from '@vueuse/core'
import { useDebounceFn } from '@vueuse/core';
// @ts-ignore
// import Guacamole from 'guacamole-common-js';
import Guacamole from '@dushixiang/guacamole-common-js';
console.log('Guacamole version:', Guacamole);
import { NSpin, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n';
import { getCurrentConnectParams } from '@/utils/common';
import { getSupportedGuacMimeTypes } from '@/utils/guacamole_helper';
import type { GuacamoleDisplay } from '@/types/guacamole.type';
import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';
import { ErrorStatusCodes, ConvertGuacamoleError } from '@/utils/status';
import { LanguageCode } from '@/locales';

const message = useMessage();
const { t } = useI18n();
const { width, height } = useWindowSize();
interface GuacamoleClient {
    sendSize: (width: number, height: number) => void;
    connect: (connectString: string) => void;
    getDisplay: () => GuacamoleDisplay;
    disconnect(): () => void;
}


const apiPrefix = ref("")
const wsPrefix = ref("")
const localCursor = ref(false);
const scale = ref(1);
const pixelDensity = window.devicePixelRatio || 1;
const sessionObject = ref()
const shareId = ref<string | null>(null);
const currentUser = ref<any | null>(null);
const onlineUsersMap = ref<Record<string, any>>({});
const warningIntervalId = ref<number | null>(null);
const actions = ref<any>();
const enableShare = ref(false);
const guacDisplay = ref<GuacamoleDisplay>();
const guacTunnel = ref(null);
const guacClient = ref<GuacamoleClient>();
const drawShow = ref(false);
const connectStatus = ref('Connecting');
const loading = ref(true);
// const keyboard = ref<Guacamole.Keyboard | null>(null);
const debouncedResize = useDebounceFn(() => {
    updateScale();
    if (guacClient.value && guacDisplay.value) {
        console.log('Sending resize to Guacamole client:', width.value, height.value);
        guacClient.value.sendSize(width.value, height.value)

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
        return 1
    }
    const newScale = Math.min(
        width.value / w,
        height.value / h
    );
    if (newScale !== scale.value) {
        console.log(`Guacamole display scaled from ${scale} to ${newScale}`);
        scale.value = newScale;
        guacDisplay.value.scale(newScale);
    }
};


watch([width, height], ([newWidth, newHeight]) => {
    debouncedResize();
}, { immediate: true });

const getConnectString = async (tokenId: string) => {
    const optimalWidth = width.value;
    const optimalHeight = height.value;
    const optimalDpi = pixelDensity * 96
    const supportMimeTypes = await getSupportedGuacMimeTypes();
    let connectString =
        'TOKEN_ID=' + encodeURIComponent(tokenId) +
        '&GUAC_WIDTH=' + Math.floor(optimalWidth) +
        '&GUAC_HEIGHT=' + Math.floor(optimalHeight) +
        '&GUAC_DPI=' + Math.floor(optimalDpi) +
        supportMimeTypes;
    return connectString;
}

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
    const dataObj = JSON.parse(data)
    switch (event) {
        case 'session_pause': {
            const msg = `${dataObj.user} ${t('PauseSession')}`
            message.info(msg)
            break
        }
        case 'session_resume': {
            const msg = `${dataObj.user} ${t('ResumeSession')}`
            message.info(msg)
            break
        }
        case 'session': {
            sessionObject.value = dataObj
            actions.value = dataObj.action_permission || {}
            enableShare.value = actions.value.enable_share || false;

            // todo: 处理 文件系统和剪切板



            //   this.session = dataObj
            //   this.initFileSystem()
            //   this.initClipboard()
            //   const actions = this.session.action_permission
            //   this.$log.debug('Session actions: ', actions)
            //   this.enableShare = actions.enable_share

            break
        }
        case 'current_user': {
            currentUser.value = dataObj
            shareId.value = dataObj.share_id || null
            break
        }
        case 'share_join': {
            if (dataObj.primary) {
                break
            }
            const joinMsg = `${dataObj.user} ${t('JoinShare')}`
            message.info(joinMsg)
            break
        }
        case 'share_exit': {
            const leaveMsg = `${dataObj.user} ${t('LeaveShare')}`
            message.info(leaveMsg)
            break
        }
        case 'share_users': {
            onlineUsersMap.value = dataObj
            break
        }
        case 'perm_expired': {
            const warningMsg = `${t('PermissionExpired')}: ${dataObj.detail}`
            message.warning(warningMsg)
            warningIntervalId.value = window.setInterval(() => {
                message.warning(warningMsg)
            }, 1000 * 31)
            break
        }
        case 'perm_valid': {
            if (warningIntervalId.value) {
                window.clearInterval(warningIntervalId.value);
                warningIntervalId.value = null;
            }
            message.success(t('PermissionValid'));
            break
        }
        default:
            break
    }

};

const onClientError = (status: any) => {
    loading.value = false;
    console.error('Guacamole client error:', status);
    const code = status.code
    let msg = status.message || t('UnknownError');
    const currentLang = LanguageCode;
    msg = ErrorStatusCodes[code] ? t(ErrorStatusCodes[code]) : t(ConvertGuacamoleError(status.message))
    console.log('Guacamole error message:', msg);
    switch (code) {
        case 1005:
            // 管理员终断会话，特殊处理
            if (currentLang === 'cn') {
                msg = status.message + ' ' + msg
            } else {
                msg = msg + ' ' + status.message
            }
            break
        case 1003:
        case 1010:
            msg = msg.replace('{PLACEHOLDER}', status.message)
            break
        case 1006:
            msg = msg + ': ' + status.message
            break
    }
    message.error(msg);

};


const connectGuacamole = async (connectString: string) => {
    const displayRef: any = document.getElementById('display');
    const tunnel = new Guacamole.WebSocketTunnel(wsPrefix.value);
    tunnel.receiveTimeout = 60 * 1000; // Set receive timeout to 60 seconds
    tunnel.onerror = (error: any) => {
        message.error(t('WebSocketError') + ` tunnel : ${error.message}`);
    };
    tunnel.onuuid = (uuid: string) => {
        tunnel.uuid = uuid;
        console.log('WebSocket UUID:', uuid);
    };
    const client = new Guacamole.Client(tunnel);
    client.onstatechange = clientStateChanged;
    client.onerror = onClientError;
    guacTunnel.value = tunnel;
    guacClient.value = client;
    const display = client.getDisplay();
    guacDisplay.value = display;
    const displayEl = display.getElement();
    displayEl.onclick = (e: any) => {
        e.preventDefault()
        return false
    }

    display.onresize = () => {
        console.log('Guacamole display onresize:', display.getWidth(), display.getHeight());
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
            mouseState.down)
        client.sendMouseState(scaledState)
    }

    const handleEmulatedMouseDown = (mouseState: any) => {
        // Emulate mouse down event
        if (client || display) { return }
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '')
        // Send mouse state, show cursor if necessary
        display.showCursor(true)
        sendScaledMouseState(mouseState)
    };
    const handleEmulatedMouseState = (mouseState: any) => {
        // Emulate mouse move/up event
        if (client || display) { return }
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.MOUSE_EVENT, '')
        // Send mouse state, hide cursor if necessary
        display.showCursor(true)
        sendScaledMouseState(mouseState)
    };

    mouse.onmousedown = mouse.onmouseup = mouse.onmousemove = (mouseState: any) => {
        // Send mouse state, hide cursor if necessary
        display.showCursor(true)
        sendScaledMouseState(mouseState)
    };
    mouse.onmouseout = (mouseState: any) => {
        // Send mouse state, hide cursor if necessary
        display.showCursor(false)
    };

    const touchScreen = new Guacamole.Mouse.Touchscreen(displayEl)
    touchScreen.onmousedown = handleEmulatedMouseDown
    touchScreen.onmousemove = touchScreen.onmouseup = handleEmulatedMouseState
    const sink = new Guacamole.InputSink()
    const guacKeyboard = new Guacamole.Keyboard(sink.getElement());
    // guacKeyboard.listenTo(sink.getElement());
    guacKeyboard.onkeydown = (keysym: any) => {
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '')
        client.sendKeyEvent(1, keysym)

    }
    guacKeyboard.onkeyup = (keysym: any) => {
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '')
        client.sendKeyEvent(0, keysym)
    }

    displayRef.appendChild(display.getElement());
    document.body.appendChild(sink.getElement());

    const handleMouseEnter = () => { if (displayEl) displayEl.style.cursor = 'none'; sink.focus() };
    const handleMouseLeave = () => { if (displayEl) displayEl.style.cursor = 'default'; };
    displayEl.addEventListener('mouseenter', handleMouseEnter);
    displayEl.addEventListener('mouseleave', handleMouseLeave);
    client.connect(connectString);
};

onMounted(async () => {
    loading.value = true;
    const handLunaOpen = (message: any) => {
        console.log('Received Luna command:', message);
        drawShow.value = true;
    }
    lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.OPEN, handLunaOpen);
    const params = getCurrentConnectParams();
    wsPrefix.value = params.ws || '';
    apiPrefix.value = params.api || '';
    const token = params['data'].token || '';
    const connectString = await getConnectString(token);
    console.log('Connect string:', connectString);
    await connectGuacamole(connectString);

});

onUnmounted(() => {
    if (guacClient.value) {
        guacClient.value.disconnect();
    }
    lunaCommunicator.offLuna(LUNA_MESSAGE_TYPE.OPEN);
    lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.CLOSE, '');
});


</script>

<template>
    <div class="w-full h-full justify-center  flex flex-col">
        <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
            <n-spin :show="loading" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
            </n-spin>
        </div>
        <div id="display" v-show="!loading" class="w-screen h-screen"></div>
    </div>
    <n-drawer v-model:show="drawShow"
    :min-width="502"  :placement="'right'"
    resizable>
        <n-drawer-content>
        
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