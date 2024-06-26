<template>
  <el-main>
    <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="rgba(0, 0, 0, 0.8">
      <div :style="divStyle">
        <div id="display" />
      </div>
    </el-row>
  </el-main>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import i18n from '@/i18n'
import { getSupportedMimetypes } from '@/utils/image'
import { getSupportedGuacAudios } from '@/utils/audios'
import { getSupportedGuacVideos } from '@/utils/video'
import { ConvertGuacamoleError, ErrorStatusCodes } from '@/utils/status'

const pixelDensity = window.devicePixelRatio || 1
export default {
  name: 'Connection',
  props: {
    wsUrl: {
      type: String,
      required: true
    },
    readonly: {
      type: Boolean,
      default: true
    },
    params: {
      type: Object,
      required: true,
      default: () => { return {} }
    }
  },
  data() {
    return {
      displayWidth: 0,
      displayHeight: 0,
      loading: true,
      loadingText: i18n.t('Connecting') + ' ...',
      resizing: false,
      keyboard: null,
      client: null,
      display: null,
      tunnel: null
    }
  },
  computed: {
    divStyle: function() {
      return {
        width: this.displayWidth + 'px',
        height: this.displayHeight + 'px',
        margin: '0 auto'
      }
    }
  },
  mounted: function() {
    this.$log.debug(this.params)
    const urlParams = new URLSearchParams()
    for (const key in this.params) {
      urlParams.append(key, encodeURIComponent(this.params[key]))
    }
    this.getConnectString(urlParams.toString()).then(connectionParams => {
      this.connectGuacamole(connectionParams, this.wsUrl)
    }).catch(err => {
      this.$log.debug(err)
    })
  },
  methods: {
    getAutoSize() {
      const optimalWidth = window.innerWidth * pixelDensity
      const optimalHeight = window.innerHeight * pixelDensity
      return [optimalWidth, optimalHeight]
    },
    getConnectString(params) {
      // Calculate optimal width/height for display
      const [optimalWidth, optimalHeight] = this.getAutoSize()
      const optimalDpi = pixelDensity * 96
      return new Promise((resolve, reject) => {
        Promise.all([
          getSupportedMimetypes(),
          getSupportedGuacAudios(),
          getSupportedGuacVideos()
        ]).then(values => {
          // ["image/jpeg", "image/png", "image/webp"]
          const supportImages = values[0]
          const supportAudios = values[1]
          const supportVideos = values[2]
          this.displayWidth = optimalWidth
          this.displayHeight = optimalHeight
          let connectString = params +
              '&GUAC_WIDTH=' + Math.floor(optimalWidth) +
              '&GUAC_HEIGHT=' + Math.floor(optimalHeight) +
              '&GUAC_DPI=' + Math.floor(optimalDpi)
          this.$log.debug('Connect string: ', connectString)
          supportImages.forEach(function(mimetype) {
            connectString += '&GUAC_IMAGE=' + encodeURIComponent(mimetype)
          })
          supportAudios.forEach(function(mimetype) {
            connectString += '&GUAC_AUDIO=' + encodeURIComponent(mimetype)
          })
          supportVideos.forEach(function(mimetype) {
            connectString += '&GUAC_VIDEO=' + encodeURIComponent(mimetype)
          })
          resolve(connectString)
        })
      })
    },
    getPropScale() {
      const display = this.client.getDisplay()
      if (!display) {
        return
      }
      // Calculate scale to fit screen
      return Math.min(
        window.innerWidth / Math.max(display.getWidth(), 1),
        window.innerHeight / Math.max(display.getHeight(), 1)
      )
    },

    updateDisplayScale() {
      const display = this.client.getDisplay()
      if (!display) {
        return
      }

      const scale = this.getPropScale()
      if (scale === this.scale) {
        return
      }
      this.scale = scale
      this.display.scale(scale)
      this.displayWidth = display.getWidth() * scale
      this.displayHeight = display.getHeight() * scale
    },

    debounce(fn, wait) {
      let timeout = null
      return function() {
        if (timeout !== null) {
          clearTimeout(timeout)
        }
        timeout = setTimeout(fn, wait)
      }
    },

    onWindowResize() {
      // 监听 window display的变化
      if (this.client !== null) {
        this.updateDisplayScale()
        // 这里不应该发过去，协作方和监控方，不能改变
        // this.client.sendSize(optimalWidth, optimalHeight)
      }
    },

    onClientConnected() {
      this.onWindowResize()
      setTimeout(() => {
        window.addEventListener('resize', this.debounce(this.onWindowResize.bind(this), 300))
      }, 500)
      // window.onfocus = this.onWindowFocus
      window.onblur = () => {
        if (this.keyboard !== null) {
          this.keyboard.reset()
        }
      }
      const display = document.getElementById('display')
      display.appendChild(this.client.getDisplay().getElement())
      display.appendChild(this.sink.getElement())
    },

    displayResize(width, height) {
      // 监听guacamole display的变化
      this.$log.debug('Display resize: ', width, height)
      const scale = this.getPropScale()
      this.display.scale(scale)
      this.displayWidth = width * scale
      this.displayHeight = height * scale
    },
    onCursor(canvas, x, y) {
      this.localCursor = true
    },

    clientStateChanged(clientState) {
      switch (clientState) {
        // Idle
        case 0:
          this.clientState = 'IDLE'
          this.$log.debug('clientState, IDLE')
          break

          // Ignore "connecting" state
        case 1: // Connecting
          this.clientState = 'Connecting'
          this.$log.debug('clientState, Connecting')
          break

          // Connected + waiting
        case 2:
          this.clientState = 'Connected + waiting'
          this.$log.debug('clientState, Connected + waiting')
          break

          // Connected
        case 3:
          this.clientState = 'Connected'
          this.$log.debug('clientState, Connected ')
          this.loading = false
          // Send any clipboard data already provided
          // if (managedClient.clipboardData)
          //     ManagedClient.setClipboard(managedClient, managedClient.clipboardData);
          //
          // Begin streaming audio input if possible
          var AUDIO_INPUT_MIMETYPE = 'audio/L16;rate=44100,channels=2'
          var requestAudioStream = function requestAudioStream(client) {
            // Create new audio stream, associating it with an AudioRecorder
            const stream = client.createAudioStream(AUDIO_INPUT_MIMETYPE)
            const recorder = Guacamole.AudioRecorder.getInstance(stream, AUDIO_INPUT_MIMETYPE)

            // If creation of the AudioRecorder failed, simply end the stream
            // eslint-disable-next-line brace-style
            if (!recorder) { stream.sendEnd() }
            // Otherwise, ensure that another audio stream is created after this
            // audio stream is closed
            else { recorder.onclose = requestAudioStream.bind(this, client) }
          }
          requestAudioStream(this.client)
          this.onClientConnected()
          break

          // Update history when disconnecting
        case 4: // Disconnecting
        case 5: // Disconnected
          this.clientState = 'Disconnecting'
          this.$log.debug('clientState, Disconnected ')
          // this.closeDisplay('clientState Disconnecting')
          var display = document.getElementById('display')
          display.innerHTML = ''
          break
      }
    },
    clientOnErr(status) {
      this.loading = false
      this.closeDisplay(status)
    },
    closeDisplay(status) {
      this.$log.debug(status, i18n.locale)
      const code = status.code
      let msg = status.message
      msg = ErrorStatusCodes[code] ? this.$t(ErrorStatusCodes[code]) : this.$t(ConvertGuacamoleError(status.message))
      this.$alert(msg, this.$t('ErrTitle'), {
        confirmButtonText: this.$t('OK'),
        callback: action => {
          const display = document.getElementById('display')
          if (this.client) {
            display.innerHTML = ''
          }
        }
      })
    },
    onMouseDown(mouseState) {
      document.body.focus()
      this.handleMouseState(mouseState)
      this.isMenuCollapse = true
      this.sink.focus()
    },

    onMouseOut(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },
    handleMouseState(mouseState) {
      if (this.readonly) { return }
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }
      // Send mouse state, show cursor if necessary
      this.display.showCursor(!this.localCursor)
      this.sendScaledMouseState(mouseState)
    },
    handleEmulatedMouseDown(mouseState) {
      this.handleEmulatedMouseState(mouseState)
      this.isMenuCollapse = true
      this.sink.focus()
      this.$log.debug('handleEmulatedMouseDown', mouseState)
    },
    handleEmulatedMouseState(mouseState) {
      if (this.readonly) { return }
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }

      // Ensure software cursor is shown
      this.display.showCursor(true)
      this.$log.debug('handleEmulatedMouseState', mouseState)
      // Send mouse state, ensure cursor is visible
      this.sendScaledMouseState(mouseState)
    },
    sendScaledMouseState(mouseState) {
      const scaledState = new Guacamole.Mouse.State(
        mouseState.x / this.display.getScale(),
        mouseState.y / this.display.getScale(),
        mouseState.left,
        mouseState.middle,
        mouseState.right,
        mouseState.up,
        mouseState.down)
      this.client.sendMouseState(scaledState)
    },
    setDisplayCallback(display) {
      display.onresize = this.displayResize
      display.oncursor = this.onCursor

      const displayEl = display.getElement()
      this.$log.debug('Display el: ', displayEl)
      displayEl.onclick = (e) => {
        e.preventDefault()
        return false
      }
      const mouse = new Guacamole.Mouse(displayEl)
      // Ensure focus is regained via mousedown before forwarding event
      mouse.onmousedown = this.onMouseDown
      mouse.onmouseup = mouse.onmousemove = this.handleMouseState
      // Hide software cursor when mouse leaves display
      mouse.onmouseout = this.onMouseOut
      this.mouse = mouse
      // touch 触屏操作
      const outDisplay = document.getElementById('display')
      const touchScreen = new Guacamole.Mouse.Touchscreen(outDisplay)
      touchScreen.onmousedown = this.handleEmulatedMouseDown
      touchScreen.onmousemove = touchScreen.onmouseup = this.handleEmulatedMouseState
      // 输入下沉
      const sink = new Guacamole.InputSink()
      sink.focus()
      this.sink = sink

      // Keyboard
      const keyboard = new Guacamole.Keyboard(sink.getElement())
      keyboard.onkeydown = (keysym) => {
        this.client.sendKeyEvent(1, keysym)
      }
      keyboard.onkeyup = (keysym) => {
        this.client.sendKeyEvent(0, keysym)
      }
      this.keyboard = keyboard
    },
    connectGuacamole(connectionParams, wsURL) {
      const display = document.getElementById('display')
      const tunnel = new Guacamole.WebSocketTunnel(wsURL)
      const client = new Guacamole.Client(tunnel)
      const vm = this
      tunnel.receiveTimeout = 60 * 1000
      tunnel.onerror = function tunnelError(status) {
        vm.$log.debug('tunnel error ', status)
        display.innerHTML = ''
      }
      tunnel.onuuid = function tunnelAssignedUUID(uuid) {
        vm.$log.debug('tunnel assigned uuid ', uuid)
        tunnel.uuid = uuid
      }
      tunnel.onstatechange = this.onTunnelStateChanged
      this.client = client
      this.tunnel = tunnel
      this.display = this.client.getDisplay()
      this.setDisplayCallback(this.display)
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      client.connect(connectionParams)

      window.onunload = function() {
        client.disconnect()
      }
    }
  }
}
</script>

<style scoped>
.el-container {
  margin: 0 auto;
}

.el-main {
  padding: 0;
}
</style>
