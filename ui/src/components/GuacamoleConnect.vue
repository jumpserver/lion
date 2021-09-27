<template>
  <el-container>
    <el-main>
      <el-row
        v-loading="loading"
        :element-loading-text="loadingText"
        element-loading-background="#1f1b1b"
      >
        <div :style="containerStyle">
          <div id="displayOuter">
            <div id="displayMiddle">
              <div id="display" />
            </div>
          </div>
        </div>
      </el-row>
    </el-main>
    <el-aside width="60px" center>
      <el-menu
        v-if="!loading"
        :collapse="true"
        background-color="#1f1b1b"
        text-color="#ffffff"
      >
        <el-menu-item :disabled="menuDisable || !clipboardInited" index="1" @click="toggleClipboard">
          <i class="el-icon-document-copy" /><span slot="title">{{ $t('Clipboard') }}</span>
        </el-menu-item>
        <el-menu-item v-if="hasFileSystem" :disabled="menuDisable" index="2" @click="toggleFileSystem">
          <i class="el-icon-folder" /><span slot="title">{{ $t('Files') }}</span>
        </el-menu-item>
        <el-submenu :disabled="menuDisable" index="1" popper-class="sidebar-popper" @mouseenter="()=>{}">
          <template slot="title">
            <i class="el-icon-position" /><span slot="title">{{ $t('Shortcuts') }}</span>
          </template>
          <el-menu-item
            v-for="(item, i) in combinationKeys"
            :key="i"
            :index="menuIndex('3-',i)"
            @click="handleKeys(item.keys)"
          >
            {{ item.name }}
          </el-menu-item>
        </el-submenu>
      </el-menu>
      <GuacClipboard
        v-if="clipboardInited"
        ref="clipboard"
        :visible.sync="clipboardDrawer"
        :client="client"
        :tunnel="tunnel"
        @closeDrawer="onCloseDrawer"
      />
      <GuacFileSystem
        v-if="fileSystemInited"
        ref="fileSystem"
        :client="client"
        :tunnel="tunnel"
        :show.sync="fileDrawer"
        @closeDrawer="onCloseDrawer"
      />
    </el-aside>
  </el-container>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { getSupportedMimetypes } from '@/utils/image'
import { getSupportedGuacAudios } from '@/utils/audios'
import { getSupportedGuacVideos } from '@/utils/video'
import { getCurrentConnectParams } from '@/utils/common'
import { createSession, deleteSession } from '@/api/session'
import GuacClipboard from './GuacClipboard'
import GuacFileSystem from './GuacFileSystem'
import { default as i18n, getLanguage } from '@/i18n'
import { ErrorStatusCodes } from '@/utils'
import { localStorageGet } from '@/utils/common'

const pixelDensity = window.devicePixelRatio || 1
const sideWidth = 60
export default {
  name: 'GuacamoleConnect',
  components: {
    GuacClipboard,
    GuacFileSystem
  },
  data() {
    return {
      apiPrefix: '/api',
      wsPrefix: '/lion/ws/connect/',
      dialogFormVisible: false,
      requireParams: [],
      isMenuCollapse: true,
      hasFileSystem: false,
      clipboardDrawer: false,
      fileDrawer: false,
      loading: true,
      session: null,
      tunnelState: '',
      fileSystemInited: false,
      clipboardInited: false,
      loadingText: i18n.t('Connecting') + ' ...',
      clientState: 'Connecting',
      localCursor: false,
      client: null,
      clientProperties: {
        autoFit: true
      },
      tunnel: null,
      displayWidth: window.innerWidth - sideWidth,
      displayHeight: window.innerHeight,
      connected: false,
      clipboardData: {
        type: 'text/plain',
        data: ''
      },
      sink: null,
      keyboard: null,
      combinationKeys: [
        {
          keys: ['65507', '65513', '65535'],
          name: 'Ctrl+Alt+Delete'
        },
        {
          keys: ['65507', '65513', '65288'],
          name: 'Ctrl+Alt+Backspace'
        },
        {
          keys: ['65515', '100'],
          name: 'Windows+D'
        },
        {
          keys: ['65515', '101'],
          name: 'Windows+E'
        },
        {
          keys: ['65515', '114'],
          name: 'Windows+R'
        },
        {
          keys: ['65515', '120'],
          name: 'Windows+X'
        },
        {
          keys: ['65515'],
          name: 'Windows'
        }
      ],
      scale: 1,
      timeout: null
    }
  },
  computed: {
    containerStyle() {
      return {
        height: this.displayHeight + 'px',
        width: this.displayWidth + 'px'
      }
    },
    menuDisable: function() {
      return !(this.clientState === 'Connected') || !(this.tunnelState === 'OPEN')
    }
  },
  mounted: function() {
    const result = getCurrentConnectParams()
    this.apiPrefix = result['api']
    this.wsPrefix = result['ws']
    const vm = this
    createSession(result['api'], result['data']).then(res => {
      window.addEventListener('beforeunload', e => this.beforeunloadFn(e))
      window.addEventListener('unload', e => this.beforeunloadFn(e))
      this.session = res.data
      this.startConnect()
    }).catch(err => {
      vm.$log.debug('err ', err.message)
    })
  },
  methods: {
    toggleFileSystem(e) {
      if (this.menuDisable) {
        return
      }
      this.fileDrawer = !this.fileDrawer
    },
    initFileSystem() {
      this.fileSystemInited = true
      setTimeout(() => {
        const dropbox = document.getElementById('display')
        dropbox.addEventListener('dragenter', function(e) {
          e.stopPropagation()
          e.preventDefault()
        }, false)
        dropbox.addEventListener('dragover', function(e) {
          e.stopPropagation()
          e.preventDefault()
        }, false)
        dropbox.addEventListener('drop', this.$refs.fileSystem.fileDrop, false)
      }, 300)
    },
    initClipboard() {
      if (this.session && this.session.permission) {
        const actions = this.session.permission.actions
        let hasClipboardPermission = false
        const clipboardActions = ['all', 'clipboard_copy',
          'clipboard_paste', 'clipboard_copy_paste']
        for (let i = 0; i < actions.length; i++) {
          if (clipboardActions.includes(actions[i])) {
            hasClipboardPermission = true
            break
          }
        }
        this.$log.debug(this.session.permission)
        this.clipboardInited = hasClipboardPermission
      }
    },
    beforeunloadFn(e) {
      this.removeSession()
    },
    startConnect() {
      this.getConnectString(this.session.id).then(connectionParams => {
        this.connectGuacamole(connectionParams, this.wsPrefix)
      })
    },
    onClientConnected() {
      this.onWindowResize()
      window.addEventListener('resize', this.debounceWindowResize)
      window.onfocus = this.onWindowFocus
    },
    removeSession() {
      deleteSession(this.apiPrefix, this.session.id).catch(err => {
        this.$log.debug(err)
      })
    },
    menuIndex(index, num) {
      return index + num
    },
    toggleClipboard() {
      if (this.menuDisable) {
        return
      }
      this.clipboardDrawer = !this.clipboardDrawer
    },

    getAutoSize() {
      const width = this.displayWidth
      const height = this.displayHeight
      this.$log.debug('auto size:', width, height)
      return [width, height]
    },

    getGuaSize() {
      const lunaSetting = localStorageGet('LunaSetting') || {}
      const solution = lunaSetting['rdpResolution']
      if (!solution || solution.toLowerCase() === 'auto' || solution.indexOf('x') === -1) {
        this.$log.debug('Solution invalid: ', solution)
        return this.getAutoSize()
      }
      let [width, height] = solution.split('x')
      width = parseInt(width)
      height = parseInt(height)
      if (isNaN(width) || width < 100 || isNaN(height) || height < 100) {
        this.$log.debug('Solution invalid2: ', solution)
        return this.getAutoSize()
      }
      return [width, height]
    },

    getConnectString(sessionId) {
      // Calculate optimal width/height for display
      const [optimalWidth, optimalHeight] = this.getGuaSize()
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
          let connectString =
              'SESSION_ID=' + encodeURIComponent(sessionId) +
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

    onTunnelStateChanged(state) {
      switch (state) {
        // Connection is being established
        case Guacamole.Tunnel.State.CONNECTING:
          this.tunnelState = 'CONNECTING'
          this.$log.debug('Tunnel state change to Tunnel.State.CONNECTING ')
          break

          // Connection is established / no longer unstable
        case Guacamole.Tunnel.State.OPEN:
          this.tunnelState = 'OPEN'
          this.$log.debug('Tunnel state change to Tunnel.State.OPEN ')
          this.initFileSystem()
          this.initClipboard()
          break

          // Connection is established but misbehaving
        case Guacamole.Tunnel.State.UNSTABLE:
          this.tunnelState = 'UNSTABLE'
          this.$log.debug('Tunnel state change to Tunnel.State.UNSTABLE ')
          break

          // Connection has closed
        case Guacamole.Tunnel.State.CLOSED:
          this.tunnelState = 'CLOSED'
          this.$log.debug('Tunnel state change to Tunnel.State.CLOSED ')
          break
        default:
          this.tunnelState = 'unknown'
          this.$log.debug('Tunnel state change tounknown ', state)
          break
      }
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
            let recorder
            try {
              recorder = Guacamole.AudioRecorder.getInstance(stream, AUDIO_INPUT_MIMETYPE)
            } catch (e) {
              console.log('Get audio recorder error, ignore')
              recorder = null
            }

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
      this.$log.debug('clientOnErr', status)
      this.closeDisplay(status)
    },

    closeDisplay(status) {
      this.$log.debug(status, i18n.locale)
      const code = status.code
      let msg = status.message
      if (getLanguage() === 'cn') {
        msg = ErrorStatusCodes[code] ? this.$t(ErrorStatusCodes[code]) : status.message
      }
      this.$alert(msg, this.$t('ErrTitle'), {
        confirmButtonText: this.$t('OK'),
        callback: action => {
          const display = document.getElementById('display')
          if (this.client) {
            // display.removeChild(this.client.getDisplay().getElement())
            display.innerHTML = ''
          }
        }
      })
    },
    onCursor(canvas, x, y) {
      this.localCursor = true
    },

    handleMouseState(mouseState) {
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }

      // Send mouse state, show cursor if necessary
      this.display.showCursor(!this.localCursor)
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

    handleEmulatedMouseState(mouseState) {
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }

      // Ensure software cursor is shown
      this.display.showCursor(true)
      this.$log.debug('handleEmulatedMouseState', mouseState)
      // Send mouse state, ensure cursor is visible
      this.sendScaledMouseState(mouseState)
    },

    handleEmulatedMouseDown(mouseState) {
      this.handleEmulatedMouseState(mouseState)
      this.isMenuCollapse = true
      this.sink.focus()
      this.$log.debug('handleEmulatedMouseDown', mouseState)
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

    onCloseDrawer() {
      this.$log.debug('onCloseDrawer', this.sink)
      this.sink.focus()
    },

    getPropScale() {
      const display = this.client.getDisplay()
      if (!display) {
        return
      }
      const [width, height] = this.getAutoSize()
      // Calculate scale to fit screen
      return Math.min(
        width / Math.max(display.getWidth(), 1),
        height / Math.max(display.getHeight(), 1)
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
      this.$log.debug('this scale', scale)
      this.display.scale(scale)
    },

    debounceWindowResize() {
      if (this.timeout) {
        clearTimeout(this.timeout)
      }
      this.timeout = setTimeout(() => this.onWindowResize(), 300)
    },

    onWindowResize() {
      // 监听 window display的变化
      this.displayWidth = window.innerWidth - sideWidth
      this.displayHeight = window.innerHeight
      const [optimalWidth, optimalHeight] = this.getGuaSize()
      if (this.client !== null) {
        const display = this.client.getDisplay()
        const displayHeight = display.getHeight() * pixelDensity
        const displayWidth = display.getWidth() * pixelDensity
        this.updateDisplayScale()
        if (displayHeight === optimalWidth && displayWidth === optimalHeight) {
          return
        }
        this.client.sendSize(optimalWidth, optimalHeight)
      }
    },

    displayResize(width, height) {
      // 监听guacamole display的变化
      this.$log.debug('Display resize: ', width, height)
      const scale = this.getPropScale()
      this.display.scale(scale)
    },

    onWindowFocus() {
      this.$log.debug('On window focus ')
      if (this.$refs.clipboard) {
        this.$refs.clipboard.sendClipboardToRemote()
      }
    },

    onsync: function(timestamp) {
      // this.$log.debug('onsync==> ', timestamp)
    },

    handleKeys(keys) {
      if (!this.client) {
        return
      }
      for (let i = 0; i < keys.length; i++) {
        this.client.sendKeyEvent(1, keys[i])
      }
      for (let i = 0; i < keys.length; i++) {
        this.client.sendKeyEvent(0, keys[i])
      }
    },

    setTunnelCallback(tunnel) {
      const vm = this
      tunnel.onerror = (status) => {
        vm.$message.error(vm.$t('WebSocketError'))
        vm.$log.error('Tunnel error: ', status)
      }
      tunnel.onuuid = (uuid) => {
        vm.$log.debug('Tunnel assigned UUID: ', uuid)
        vm.tunnel.uuid = uuid
      }
      tunnel.onstatechange = vm.onTunnelStateChanged
    },

    onFileSystem(obj, name) {
      this.hasFileSystem = true
      if (this.$refs.fileSystem) { this.$refs.fileSystem.fileSystemReceived(obj, name) }
    },

    setClientCallback(client) {
      const vm = this
      // client.onrequired = this.onRequireParams
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      // 文件挂载
      client.onfilesystem = this.onFileSystem
      client.onfile = (stream, mimetype, filename) => {
        return vm.$refs.fileSystem.clientFileReceived(stream, mimetype, filename)
      }
      // 剪贴板
      client.onclipboard = (stream, mimetype) => {
        return vm.$refs.clipboard.receiveClientClipboard(stream, mimetype)
      }
      client.onsync = this.onsync
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
      const displayRef = document.getElementById('display')
      const tunnel = new Guacamole.WebSocketTunnel(wsURL)
      // 连接资产耗时，造成的 ws 超时断开问题 默认 15s 改成 60s
      tunnel.receiveTimeout = 60 * 1000
      const client = new Guacamole.Client(tunnel)
      this.client = client
      this.tunnel = tunnel
      this.display = client.getDisplay()

      this.setTunnelCallback(this.tunnel)
      this.setClientCallback(this.client)
      this.setDisplayCallback(this.display)
      // client.getDisplay()
      displayRef.appendChild(this.display.getElement())
      displayRef.appendChild(this.sink.getElement())
      this.$log.debug('Display : ', displayRef)
      // 开始连接
      client.connect(connectionParams)
      window.onunload = function() {
        client.disconnect()
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.el-container {
  margin: 0 auto;
}

.el-main {
  padding: 0;
  overflow: hidden;
  width: 100%;
  height: 100%;
}
.el-row {
  height: 100%;
}

.el-aside {
  height: 100%;
  overflow: hidden;
}

.el-dropdown-link {
  color: #409eff;
}

.el-icon-arrow-down {
  font-size: 12px;
}
.el-menu {
  height: 100%;
}

.el-menu-item {
  color: white;
  background-color: rgb(60, 56, 56);
}

.el-menu-item:hover {
  background-color: #463e3e;
}

#displayOuter {
  height: 100%;
  width: 100%;
  position: absolute;
  left: 0;
  top: 0;
  display: table;
}

#displayMiddle {
  width: 100%;
  height: 100%;
  display: table-cell;
  vertical-align: middle;
  text-align: center;
}

#display{
  display: inline-block;
}
#display * {
  position: relative;
}

#display > * {
  margin-left: auto;
  margin-right: auto;
}
</style>
