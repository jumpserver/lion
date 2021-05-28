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
      menu-trigger="click"
      :collapse-transition="false"
      @click.native="isMenuCollapse = false"
    >
      <el-menu-item :disabled="menuDisable" index="2">
        <i class="el-icon-document-copy" /><span @click="toggleClipboard">{{ $t('Clipboard') }}</span>
      </el-menu-item>
      <el-menu-item v-if="hasFileSystem" :disabled="menuDisable" index="3" @click="toggleFileSystem">
        <i class="el-icon-folder" /><span>{{ $t('Files') }}</span>
      </el-menu-item>
      <el-submenu :disabled="menuDisable" @mouseenter="()=>{}" index="1" popper-class="sidebar-popper">
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
    </el-menu>
    <el-dialog :title="$t('RequireParams')" :visible="dialogFormVisible" @close="cancelSubmitParams">
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
        <el-button type="primary" @click="submitParams">{{ $t('OK') }}</el-button>
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

    <el-dialog
      :visible.sync="manualDialogVisible"
      center
    >
      <el-form :model="manualForm">
        <el-form-item :label="$t('Username')">
          <el-input v-model="manualForm.username" autocomplete="off" />
        </el-form-item>
        <el-form-item :label="$t('Password')">
          <el-input v-model="manualForm.password" show-password autocomplete="off" />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="manualCancelClick">{{ $t('Cancel') }}</el-button>
        <el-button type="primary" :disabled="disableSubmit" @click="manualSubmitClick">{{ $t('Submit') }}</el-button>
        <el-button type="primary" @click="manualSkipClick">{{ $t('Skip') }}</el-button>
      </span>
    </el-dialog>
  </el-container>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { getSupportedMimetypes } from '@/utils/image'
import { getSupportedGuacAudios } from '@/utils/audios'
import { getSupportedGuacVideos } from '@/utils/video'
import { getCurrentConnectParams } from '@/utils/common'
import { createSession, deleteSession, updateSession } from '@/api/session'
import GuacClipboard from './GuacClipboard'
import GuacFileSystem from './GuacFileSystem'
import i18n from '@/i18n'
import { ErrorStatusCodes } from '@/utils/status'
import { getLanguage } from '../i18n'

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
      ],
      manualForm: {
        username: '',
        password: ''
      },
      manualDialogVisible: false
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
    },
    disableSubmit: function() {
      return (this.manualForm.username === '') || (this.manualForm.password === '')
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
      if (this.checkIsManualLogin(res.data)) {
        this.$log.debug('manual login', res.data)
        this.manualForm.username = res.data.system_user.username
        this.manualDialogVisible = true
        this.$log.debug(this.manualForm)
        return
      }
      this.startConnect()
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
      this.requireParams = []
    },
    beforeunloadFn(e) {
      this.removeSession()
    },
    checkIsManualLogin(session) {
      return session.login_mode === 'manual'
    },
    manualCancelClick() {
      this.$log.debug('manual cancel click')
      this.manualDialogVisible = false
      this.removeSession()
    },
    manualSubmitClick() {
      this.$log.debug('manual submit click')
      updateSession(this.apiPrefix, this.session.id, this.manualForm).then(data => {
        this.manualDialogVisible = false
        this.startConnect()
      }).catch(err => {
        this.$log.debug(err)
        this.removeSession()
      })
    },
    manualSkipClick() {
      this.$log.debug('manual skip click')
      this.manualDialogVisible = false
      this.startConnect()
    },
    startConnect() {
      window.addEventListener('resize', this.onWindowResize)
      window.onfocus = this.onWindowFocus
      this.getConnectString(this.session.id).then(connectionParams => {
        this.connectGuacamole(connectionParams, this.wsPrefix)
      })
    },
    removeSession() {
      deleteSession(this.apiPrefix, this.session.id).catch(err => {
        this.$log.debug(err)
      })
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
      const optimalDpi = pixelDensity * 96
      const optimalWidth = window.innerWidth * pixelDensity - 32
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
              '&GUAC_DPI=' + Math.floor(optimalDpi)
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
      this.client.sendMouseState(mouseState, true)
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

    onWindowResize() {
      // 监听 window display的变化
      const pixelDensity = window.devicePixelRatio || 1
      const optimalWidth = window.innerWidth * pixelDensity
      const optimalHeight = window.innerHeight * pixelDensity
      const width = optimalWidth - 32
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
      client.onrequired = this.onRequireParams
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
  overflow: hidden;
}

.el-dropdown-link {
  color: #409eff;
}

.el-icon-arrow-down {
  font-size: 12px;
}

.el-menu {
  background-color: rgb(60, 56, 56);
  border: solid 1px rgb(60, 56, 56);

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
