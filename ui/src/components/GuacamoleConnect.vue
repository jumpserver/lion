<template>
  <el-container>
    <el-main>
      <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="#676a6c">
        <div :style="divStyle">
          <div id="display" />
        </div>
      </el-row>
    </el-main>
    <div />
    <el-menu
      v-if="!loading"
      :collapse="isMenuCollapse"
      class="menu"
      @mouseover.native="isMenuCollapse = false"
    >
      <el-submenu :disabled="menuDisable" index="1">
        <template slot="title">
          <i class="el-icon-position" /><span>{{ $t('Shortcuts') }}</span>
        </template>
        <el-menu-item
          v-for="(item, i) in combinationKeys"
          :key="i"
          :index="menuIndex('1-',i)"
          @click="handleKeys(item.keys)"
        >
          {{ item.name }}
        </el-menu-item>
      </el-submenu>
      <el-menu-item :disabled="menuDisable" index="2">
        <i class="el-icon-document-copy" /><span @click="toggleClipboard">{{ $t('Clipboard') }}</span>
      </el-menu-item>
      <el-menu-item :disabled="menuDisable" index="3" @click="toggleFileSystem">
        <i class="el-icon-folder" /><span>{{ $t('Files') }}</span>
      </el-menu-item>
    </el-menu>
    <el-dialog title="认证参数" :visible="dialogFormVisible" @close="cancelSubmitParams">
      <el-form label-position="left" label-width="80px" @submit.native.prevent="submitParams">
        <el-form-item v-for="(item, index) in requireParams" :key="index" :label="item.name">
          <template v-if="checkPasswordInput(item.name)">
            <el-input v-model="item.value" show-password />
          </template>
          <template v-else>
            <el-input v-model="item.value" />
          </template>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitParams">确 定</el-button>
      </div>
    </el-dialog>
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
  </el-container>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { getSupportedMimetypes } from '@/utils/image'
import { getSupportedGuacAudios } from '@/utils/audios'
import { getSupportedGuacVideos } from '@/utils/video'
import { getCurrentConnectParams } from '@/utils/common'
import { createSession } from '@/api/session'
import GuacClipboard from './GuacClipboard'
import GuacFileSystem from './GuacFileSystem'
import i18n from '@/i18n'

export default {
  name: 'GuacamoleConnect',
  components: {
    GuacClipboard,
    GuacFileSystem
  },
  data() {
    return {
      apiPrefix: '/api',
      dialogFormVisible: false,
      requireParams: [],
      isMenuCollapse: true,
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
      tunnel: null,
      displayWidth: 0,
      displayHeight: 0,
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
      ]
    }
  },
  computed: {
    divStyle: function() {
      return {
        width: this.displayWidth + 'px',
        height: this.displayHeight + 'px'
      }
    },
    menuDisable: function() {
      return !(this.clientState === 'Connected') || !(this.tunnelState === 'OPEN')
    }
  },
  mounted: function() {
    const result = getCurrentConnectParams()
    this.apiPrefix = result['api']
    const vm = this
    createSession(result['api'], result['data']).then(res => {
      this.session = res.data
      window.addEventListener('resize', this.onWindowResize)
      window.onfocus = this.onWindowFocus
      this.getConnectString(res.data.id).then(connectionParams => {
        this.connectGuacamole(connectionParams, result['ws'])
      })
    }).catch(err => {
      vm.$log.debug('err ', err.message)
    })
  },
  methods: {
    checkPasswordInput(name) {
      return name.match('password')
    },
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
      this.clipboardInited = true
    },
    submitParams() {
      if (this.client) {
        for (let i = 0; i < this.requireParams.length; i++) {
          const stream = this.client.createArgumentValueStream('text/plain', this.requireParams[i].name)
          const writer = new Guacamole.StringWriter(stream)
          writer.sendText(this.requireParams[i].value)
          writer.sendEnd()
        }
      }
      this.dialogFormVisible = false
      this.requireParams = []
    },
    cancelSubmitParams() {
      this.dialogFormVisible = false
      if (this.client) {
        // this.client.disconnect()
      }
      this.requireParams = []
    },
    onRequireParams(params) {
      this.requireParams = []
      for (let i = 0; i < params.length; i++) {
        this.requireParams.push({
          name: params[i],
          value: ''
        })
      }
      this.dialogFormVisible = true
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

    getConnectString(sessionId) {
      // Calculate optimal width/height for display
      const pixelDensity = window.devicePixelRatio || 1
      const optimal_dpi = pixelDensity * 96
      const optimalWidth = window.innerWidth * pixelDensity - 30
      const optimalHeight = window.innerHeight * pixelDensity
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
          var connectString =
              'SESSION_ID=' + encodeURIComponent(sessionId) +
              '&GUAC_WIDTH=' + Math.floor(optimalWidth) +
              '&GUAC_HEIGHT=' + Math.floor(optimalHeight) +
              '&GUAC_DPI=' + Math.floor(optimal_dpi)
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
            var stream = client.createAudioStream(AUDIO_INPUT_MIMETYPE)
            var recorder = Guacamole.AudioRecorder.getInstance(stream, AUDIO_INPUT_MIMETYPE)

            // If creation of the AudioRecorder failed, simply end the stream
            // eslint-disable-next-line brace-style
            if (!recorder) { stream.sendEnd() }
            // Otherwise, ensure that another audio stream is created after this
            // audio stream is closed
            else { recorder.onclose = requestAudioStream.bind(this, client) }
          }
          requestAudioStream(this.client)
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

    clientOnErr(stats) {
      this.closeDisplay(stats)
    },

    closeDisplay(stats) {
      this.$log.debug('Close display, stats: ', stats)
      this.$alert('关闭窗口=== ' + stats.message, stats, {
        confirmButtonText: this.$t('Confirm'),
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
      this.client.sendMouseState(mouseState, true)
    },

    onMouseDown(mouseState) {
      document.body.focus()
      this.handleMouseState(mouseState)
      this.isMenuCollapse = true
    },

    onMouseOut(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },

    onCloseDrawer() {
      this.$log.debug('onCloseDrawer', this.sink)
      this.sink.focus()
    },

    onWindowResize() {
      // 监听 window display的变化
      const pixelDensity = window.devicePixelRatio || 1
      const optimalWidth = window.innerWidth * pixelDensity
      const optimalHeight = window.innerHeight * pixelDensity
      const width = optimalWidth - 30
      const height = optimalHeight
      if (this.client !== null) {
        const display = this.client.getDisplay()
        const displayHeight = display.getHeight() * pixelDensity
        const displayWidth = display.getWidth() * pixelDensity
        if (displayHeight === width && displayWidth === height) {
          return
        }
        this.client.sendSize(width, height)
      }
    },

    displayResize(width, height) {
      // 监听guacamole display的变化
      this.$log.debug('Display resize: ', width, height)
      this.displayWidth = width
      this.displayHeight = height
    },

    onWindowFocus() {
      this.$log.debug('On window focus ')
      this.$refs.clipboard.sendClipboardToRemote()
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
        vm.logger.error('Tunnel error: ', status)
      }
      tunnel.onuuid = (uuid) => {
        vm.$log.debug('Tunnel assigned UUID: ', uuid)
        vm.tunnel.uuid = uuid
      }
      tunnel.onstatechange = vm.onTunnelStateChanged
    },

    setClientCallback(client) {
      const vm = this
      client.onrequired = this.onRequireParams
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      // 文件挂载
      client.onfilesystem = (obj, name) => {
        return vm.$refs.fileSystem.fileSystemReceived(obj, name)
      }
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
}

.el-dropdown-link {
  color: #409eff;
}

.el-icon-arrow-down {
  font-size: 12px;
}

.el-menu {
  background-color: rgb(60, 56, 56);
  border: none;

  /deep/ .el-submenu {
    background-color: rgb(60, 56, 56);

    .el-submenu__title {
      color: white;
    }
    .el-menu-item {
      line-height: 36px;
      height: 36px;
    }
    .el-submenu__title:hover {
      background-color: #463e3e;
    }
  }

  /deep/ .el-submenu:hover {
    background-color: #463e3e;
  }
}

.el-menu-item {
  color: white;
  background-color: rgb(60, 56, 56);
  padding-left: 20px;
}

.el-menu-item:hover {
  background-color: #463e3e;
}

.el-menu--collapse {
  width: 32px;

  .el-menu-item {
    padding-left: 5px !important;
  }

  /deep/ .el-submenu__title {
    padding-left: 5px !important;
  }
}

</style>
