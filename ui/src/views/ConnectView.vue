<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue';
import { useWindowSize } from '@vueuse/core'
// @ts-ignore
import Guacamole from 'guacamole-common-js';

import { getCurrentConnectParams } from '@/utils/common';
import { getSupportedGuacMimeTypes } from '@/utils/guacamole_helper';
import type { GuacamoleDisplay } from '@/types/guacamole.type';
import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';
const { width, height } = useWindowSize();

import { useDebounceFn } from '@vueuse/core';

interface GuacamoleClient {
    sendSize: (width: number, height: number) => void;
    connect: (connectString: string) => void;
    getDisplay: () => GuacamoleDisplay;
}


const apiPrefix = ref("")
const wsPrefix = ref("")

const scale = ref(1);
const pixelDensity = window.devicePixelRatio || 1;

const guacDisplay = ref<GuacamoleDisplay>();
const guacTunnel = ref(null);
const guacClient = ref<GuacamoleClient>();
// const keyboard = ref<Guacamole.Keyboard | null>(null);
const debouncedResize = useDebounceFn((width, height) => {
    if (guacClient.value && guacDisplay.value) {
        guacClient.value.sendSize(width, height)
    }
}, 300);

const updateScale = () => {
    if (guacClient.value && guacDisplay.value) {
        const w = guacDisplay.value.getWidth();
        const h = guacDisplay.value.getHeight();

        if (h === 0 || w === 0) {
            return 1
        }
        const newScale = Math.min(
            currentWidth.value / w,
            currentHeight.value / h
        );
        if (newScale !== scale.value) {
            scale.value = newScale;
            guacDisplay.value.scale(newScale);
        }
        console.log('Guacamole display scaled to:', currentWidth.value,
            currentHeight.value, h, w,
            newScale);
    }
};



const currentWidth = ref(width)
const currentHeight = ref(height)
watch([width, height], ([newWidth, newHeight]) => {
    currentWidth.value = newWidth;
    currentHeight.value = newHeight;
    console.log('Window size changed:', newWidth, newHeight);
    //  updateScale();
    debouncedResize(newWidth, newHeight);
}, { immediate: true });

const getConnectString = async (tokenId: string) => {
    const optimalWidth = currentWidth.value || 1280;
    const optimalHeight = currentHeight.value || 720;
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



const localCursor = ref(false);
const connectGuacamole = (connectString: string) => {
    const displayRef: any = document.getElementById('display');
    console.log('Connecting to Guacamole with connect string:', wsPrefix.value);
    const tunnel = new Guacamole.WebSocketTunnel(wsPrefix.value);
    tunnel.receiveTimeout = 60 * 1000; // Set receive timeout to 60 seconds
    const client = new Guacamole.Client(tunnel);
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
        console.log('Guacamole display resized:', display.getWidth(), display.getHeight());
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
        console.log('Key down event:', keysym);
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '')
        client.sendKeyEvent(1, keysym)

    }
    guacKeyboard.onkeyup = (keysym: any) => {
        lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.KEYBOARDEVENT, '')
        client.sendKeyEvent(0, keysym)
    }

    displayRef.appendChild(display.getElement());
    document.body.appendChild(sink.getElement());

    const handleMouseEnter = () => { if (displayEl) displayEl.style.cursor = 'none'; sink.focus()};
    const handleMouseLeave = () => { if (displayEl) displayEl.style.cursor = 'default'; };
    displayEl.addEventListener('mouseenter', handleMouseEnter);
    displayEl.addEventListener('mouseleave', handleMouseLeave);
    client.connect(connectString);
};

onMounted(async () => {
    const handLunaOpen = (message: any) => {
        console.log('Received Luna command:', message);
    }
    lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.OPEN, handLunaOpen);
    const params = getCurrentConnectParams();
    wsPrefix.value = params.ws || '';
    apiPrefix.value = params.api || '';
    console.log('Connect params:', params);
    const token = params['data'].token || '';
    const connectString = await getConnectString(token);
    console.log('Connect string:', connectString);
    connectGuacamole(connectString);

});

onUnmounted(() => {
    console.log('Unmounting ConnectView, cleaning up resources');
    if (guacClient.value) {
        guacClient.value.disconnect();
    }
    lunaCommunicator.offLuna(LUNA_MESSAGE_TYPE.OPEN);

});

</script>


<template>
    <div class="w-full h-full justify-center">
        <div id="display">
        </div>
    </div>
  
</template>

<style scoped>
#display {
    width: 100vw;
    height: 100vh;
    position: relative;
}
</style>