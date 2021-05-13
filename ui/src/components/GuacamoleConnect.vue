<template>
  <el-container>
    <el-main>
      <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="rgba(0, 0, 0, 0.8">
        <div :style="divStyle">
          <div id="display" />
        </div>
      </el-row>
    </el-main>
    <div />
    <el-menu
      :collapse="isMenuCollapse"
      @mouseover.native="isMenuCollapse = false"
    >
      <el-submenu :disabled="menuDisable" index="1">
        <template slot="title">
          <i class="el-icon-position" /><span>快捷键</span>
        </template>
        <el-menu-item v-for="(item, i) in combinationKeys" :key="i" :index="menuIndex('1-',i)" @click="handleKeys(item.keys)">
          {{ item.name }}
        </el-menu-item>
      </el-submenu>
      <el-menu-item :disabled="menuDisable" index="2">
        <i class="el-icon-document-copy" /><span @click="toggleClipboard">剪切板</span>
      </el-menu-item>
      <el-menu-item :disabled="menuDisable" index="3" @click="toggleFileSystem">
        <i class="el-icon-folder" /><span>文件管理</span>
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
      loadingText: '连接中。。。',
      clientState: '连接中。。。',
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
    createSession(result['api'], result['data']).then(res => {
      this.session = res.data
      window.addEventListener('resize', this.onWindowResize)
      window.onfocus = this.onWindowFocus
      this.getConnectString(res.data.id).then(connectionParams => {
        this.connectGuacamole(connectionParams, result['ws'])
      })
    }).catch(err => {
      console.log('err ', err.message)
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
      console.log('Toggle file to: ', !this.fileDrawer, e)
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

        this.client.onfile = this.$refs.fileSystem.fileSystemReceived
        this.client.onfilesystem = this.$refs.fileSystem.fileSystemReceived
      }, 300)
    },
    initClipboard() {
      this.clipboardInited = true
      setTimeout(() => {
        this.client.onclipboard = this.$refs.clipboard.receiveClientClipboard
      }, 300)
    },
    submitParams() {
      if (this.client) {
        for (let i = 0; i < this.requireParams.length; i++) {
          var stream = this.client.createArgumentValueStream('text/plain', this.requireParams[i].name)
          var writer = new Guacamole.StringWriter(stream)
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
      const pixel_density = window.devicePixelRatio || 1
      const optimal_dpi = pixel_density * 96
      const optimal_width = window.innerWidth * pixel_density - 64
      const optimal_height = window.innerHeight * pixel_density
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
          this.displayWidth = optimal_width
          this.displayHeight = optimal_height
          var connectString =
              'SESSION_ID=' + encodeURIComponent(sessionId) +
              '&GUAC_WIDTH=' + Math.floor(optimal_width) +
              '&GUAC_HEIGHT=' + Math.floor(optimal_height) +
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
          console.log('tunnelStateChanged Tunnel.State.CONNECTING ')
          break

          // Connection is established / no longer unstable
        case Guacamole.Tunnel.State.OPEN:
          this.tunnelState = 'OPEN'
          console.log('tunnelStateChanged Tunnel.State.OPEN ')
          break

          // Connection is established but misbehaving
        case Guacamole.Tunnel.State.UNSTABLE:
          this.tunnelState = 'UNSTABLE'
          console.log('tunnelStateChanged Tunnel.State.UNSTABLE ')
          break

          // Connection has closed
        case Guacamole.Tunnel.State.CLOSED:
          this.tunnelState = 'CLOSED'
          console.log('tunnelStateChanged Tunnel.State.CLOSED ')
          break
        default:
          this.tunnelState = 'unknown'
          console.log('tunnelStateChanged unknown ', state)
          break
      }
    },

    clientStateChanged(clientState) {
      switch (clientState) {
        // Idle
        case 0:
          this.clientState = 'IDLE'
          console.log('clientState, IDLE')
          break

          // Ignore "connecting" state
        case 1: // Connecting
          this.clientState = 'Connecting'
          this.loadingText = 'Connecting'
          console.log('clientState, Connecting')
          break

          // Connected + waiting
        case 2:
          this.clientState = 'Connected + waiting'
          console.log('clientState, Connected + waiting')
          break

          // Connected
        case 3:
          this.clientState = 'Connected'
          console.log('clientState, Connected ')
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
            console.log(stream, recorder)
          }
          requestAudioStream(this.client)
          break

          // Update history when disconnecting
        case 4: // Disconnecting
        case 5: // Disconnected
          this.clientState = 'Disconnecting'
          console.log('clientState, Disconnected ')
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
      console.log(stats)
      this.$alert('关闭窗口=== ' + stats.message, stats, {
        confirmButtonText: '确定',
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
      // this.isMenuCollapse = true
    },

    onMouseOut(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },

    onCloseDrawer() {
      console.log('onCloseDrawer', this.sink)
      this.sink.focus()
    },

    onWindowResize() {
      // 监听 window display的变化
      const pixel_density = window.devicePixelRatio || 1
      const optimal_width = window.innerWidth * pixel_density
      const optimal_height = window.innerHeight * pixel_density
      const width = optimal_width - 64
      const height = optimal_height
      if (this.client !== null) {
        const display = this.client.getDisplay()
        const displayHeight = display.getHeight() * pixel_density
        const displayWidth = display.getWidth() * pixel_density
        if (displayHeight === width && displayWidth === height) {
          return
        }
        this.client.sendSize(width, height)
      }
    },

    displayResize(width, height) {
      // 监听guacamole display的变化
      console.log('on display ', width, height)
      this.displayWidth = width
      this.displayHeight = height
    },

    onWindowFocus() {
      console.log('onWindowFocus ')
      this.$refs.clipboard.sendClipboardToRemote()
    },

    onsync: function(timestamp) {
      // console.log('onsync==> ', timestamp)
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

    connectGuacamole(connectionParams, wsURL) {
      const display = document.getElementById('display')
      const tunnel = new Guacamole.WebSocketTunnel(wsURL)
      const client = new Guacamole.Client(tunnel)
      tunnel.onerror = function tunnelError(status) {
        this.$message.error('WebSocket 连接失败，请检查网络')
      }
      tunnel.onuuid = function tunnelAssignedUUID(uuid) {
        console.log('tunnelAssignedUUID ', uuid)
        tunnel.uuid = uuid
      }
      client.onrequired = this.onRequireParams

      tunnel.onstatechange = this.onTunnelStateChanged
      this.client = client
      this.tunnel = tunnel
      this.display = this.client.getDisplay()
      this.display.onresize = this.displayResize
      // client.getDisplay()
      display.appendChild(client.getDisplay().getElement())
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      // 处理从虚拟机收到的剪贴板内容
      this.initClipboard()

      this.initFileSystem()

      client.onsync = this.onsync
      // Handle any received files
      client.connect(connectionParams)

      window.onunload = function() {
        client.disconnect()
      }
      var mouse = new Guacamole.Mouse(client.getDisplay().getElement())
      // Ensure focus is regained via mousedown before forwarding event
      mouse.onMouseDown = this.onMouseDown

      mouse.onmouseup = mouse.onmousemove = this.handleMouseState
      // Hide software cursor when mouse leaves display
      mouse.onMouseOut = function() {
        if (!client.getDisplay()) return
        client.getDisplay().showCursor(false)
      }
      client.getDisplay().onCursor = this.onCursor
      client.getDisplay().getElement().onclick = function(e) {
        e.preventDefault()
        return false
      }
      const sink = new Guacamole.InputSink()
      display.appendChild(sink.getElement())
      sink.focus()
      // Keyboard
      const keyboard = new Guacamole.Keyboard(sink.getElement())
      keyboard.onkeydown = function(keysym) {
        client.sendKeyEvent(1, keysym)
      }
      keyboard.onkeyup = function(keysym) {
        client.sendKeyEvent(0, keysym)
      }
      this.sink = sink
      this.keyboard = keyboard
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
  cursor: pointer;
  color: #409eff;
}

.el-icon-arrow-down {
  font-size: 12px;
}
</style>
