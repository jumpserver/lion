<template>
  <el-container>
    <el-main>
      <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="rgba(0, 0, 0, 0.8">
        <div v-bind:style="divStyle">
          <div id="display"></div>
        </div>
      </el-row>
    </el-main>

    <el-menu :collapse="isMenuCollapse"
             @mouseover.native="isMenuCollapse = false"
    >
      <el-submenu :disabled="menuDisable" index="1">
        <template slot="title">
          <i class="el-icon-position"></i>
          <span>快捷键</span>
        </template>
        <el-menu-item v-for="(item, i) in combinationKeys" :key="i" v-bind:index="menuIndex('1-',i)" @click="handleKeys(item.keys)">
          {{ item.name }}
        </el-menu-item>
      </el-submenu>

      <el-menu-item :disabled="menuDisable" index="2"><i class="el-icon-document-copy"></i>
        <span @click="toggleClipboard">剪切板</span>
        <el-drawer direction="ltr" :visible.sync="clipboardDrawer" @close="onCloseDrawer">
          <el-row>
            <el-col :span="12" :offset="8">
              <div class="grid-content bg-purple">
                <GuacClipboard v-bind:value="clipboardText" v-on:ClipboardChange="onClipboardChange"/>
              </div>
            </el-col>
          </el-row>
        </el-drawer>
      </el-menu-item>
      <el-menu-item v-if="currentFilesystem.object" :disabled="menuDisable" index="3"><i class="el-icon-folder"></i>
        <span @click="toggleFile">文件管理</span>
        <el-drawer direction="ltr" :visible.sync="fileDrawer" @close="onCloseDrawer">
          <el-row>
            <div>{{ currentFilesystem.name }}</div>
            <GuacFileSystem ref="filesystem"
                            v-bind:guac-object="currentFilesystem.object"
                            v-bind:currentFolder="currentFolder"
                            v-on:ChangeFolder="onChangeFolder"
                            v-on:DownLoadReceived="onDownloadFile"
                            v-on:UploadFile="onUploadFiles"
            />
          </el-row>
        </el-drawer>
      </el-menu-item>
    </el-menu>
    <el-dialog title="认证参数" :visible="dialogFormVisible" @close="cancelSubmitParams">
      <el-form label-position="left" label-width="80px" @submit.native.prevent="submitParams">
        <el-form-item v-for="(item, index) in requireParams" :key="index" :label="item.name">
          <template v-if="checkPasswordInput(item.name)">
            <el-input v-model="item.value" show-password></el-input>
          </template>
          <template v-else>
            <el-input v-model="item.value"></el-input>
          </template>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitParams">确 定</el-button>
      </div>
    </el-dialog>
  </el-container>

</template>

<script>
import Guacamole from 'guacamole-common-js'
import {GetSupportedMimetypes} from '../utils/image'
import {BaseURL, getCurrentConnectParams, sanitizeFilename} from '../utils/common'
import {createSession} from '../api/session'
import GuacClipboard from './GuacClipboard'
import GuacFileSystem from './GuacFileSystem'

export default {
  name: 'GuacamoleConnect',
  components: {
    GuacClipboard,
    GuacFileSystem,
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
      loadingText: '连中。。。',
      clientState: '连中。。。',
      localCursor: false,
      client: null,
      tunnel: null,
      displayWidth: 0,
      displayHeight: 0,
      clipboardText: '',
      clipboardData: {
        type: 'text/plain',
        data: '',
      },
      sink: null,
      keyboard: null,
      currentFilesystem: {
        object: null,
        name: '',
      },
      currentFolder: {
        mimetype: Guacamole.Object.STREAM_INDEX_MIMETYPE,
        streamName: Guacamole.Object.ROOT_STREAM,
        type: 'DIRECTORY',
        files: {},
        parent: null,
      },
      combinationKeys: [
        {
          keys: ['65507', '65513', '65535'],
          name: 'Ctrl+Alt+Delete',
        },
        {
          keys: ['65507', '65513', '65288'],
          name: 'Ctrl+Alt+Backspace',
        },
        {
          keys: ['65515', '100'],
          name: 'Windows+D',
        },
        {
          keys: ['65515', '101'],
          name: 'Windows+E',
        },
        {
          keys: ['65515', '114'],
          name: 'Windows+R',
        },
        {
          keys: ['65515', '120'],
          name: 'Windows+X',
        },
        {
          keys: ['65515'],
          name: 'Windows',
        },
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
    let result = getCurrentConnectParams()
    this.apiPrefix = result['api']
    createSession(result['api'], result['data']).then(res => {
      this.session = res.data
      window.addEventListener('resize', this.onWindowResize)
      window.onfocus = this.onWindowFocus
      console.log(res.data)
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
          value: '',
        })
      }
      this.dialogFormVisible = true
    },

    menuIndex(index, num) {
      return index + num
    },

    onUploadFiles(files) {
      for (let i = 0; i < files.length; i++) {
        let streamName
        if (this.currentFolder) {
          streamName = this.currentFolder.streamName + '/' + files[i].name
        }
        this.loadingText = 'Upload Files'
        this.loading = true
        this.handleFiles(files[i], this.currentFilesystem.object, streamName).then(() => {
          this.loading = false
          this.loadingText = ''
          this.$refs.filesystem.refresh()
        })
      }
    },

    onDownloadFile(stream, mimetype, filename) {
      this.clientFileReceived(stream, mimetype, filename)
    },

    onChangeFolder(fileItem) {
      this.currentFolder = fileItem
    },

    onClipboardChange(data) {
      console.log('ClipboardChange emit ', data)
      this.clipboardText = data
      this.sendClientClipboard({
        'data': data,
        'type': 'text/plain'
      })
      this.setLocalClipboard(data)
    },

    toggleClipboard() {
      if (this.menuDisable) {
        return
      }
      this.clipboardDrawer = !this.clipboardDrawer
    },

    toggleFile() {
      if (this.menuDisable || !this.currentFilesystem.object) {
        return
      }
      this.fileDrawer = !this.fileDrawer
    },

    getSupportedGuacAudios() {
      return Guacamole.AudioPlayer.getSupportedTypes()
    },

    getSupportedGuacVideos() {
      return Guacamole.VideoPlayer.getSupportedTypes()
    },

    getConnectString(sessionId) {
      // Calculate optimal width/height for display
      let pixel_density = window.devicePixelRatio || 1
      let optimal_dpi = pixel_density * 96
      let optimal_width = window.innerWidth * pixel_density - 64
      let optimal_height = window.innerHeight * pixel_density
      return new Promise((resolve, reject) => {
        Promise.all([
          GetSupportedMimetypes(),
          this.getSupportedGuacAudios(),
          this.getSupportedGuacVideos()
        ]).then(values => {
          // ["image/jpeg", "image/png", "image/webp"]
          let supportImages = values[0]
          let supportAudios = values[1]
          let supportVideos = values[2]
          this.displayWidth = optimal_width
          this.displayHeight = optimal_height
          var connectString =
              'SESSION_ID=' + encodeURIComponent(sessionId)
              + '&GUAC_WIDTH=' + Math.floor(optimal_width)
              + '&GUAC_HEIGHT=' + Math.floor(optimal_height)
              + '&GUAC_DPI=' + Math.floor(optimal_dpi)
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
          this.closeDisplay('tunnelStateChanged CLOSED')
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
            if (!recorder)
              stream.sendEnd()
                // Otherwise, ensure that another audio stream is created after this
            // audio stream is closed
            else
              recorder.onclose = requestAudioStream.bind(this, client)
            console.log(stream, recorder)
          }
          requestAudioStream(this.client)
          break

          // Update history when disconnecting
        case 4: // Disconnecting
        case 5: // Disconnected
          this.clientState = 'Disconnecting'
          console.log('clientState, Disconnected ')
          this.closeDisplay('clientState Disconnecting')
          break

      }
    },

    clientOnErr(stats) {
      this.closeDisplay(stats)
    },

    closeDisplay(stats) {
      console.log(stats)
      this.$alert('关闭窗口===', stats, {
        confirmButtonText: '确定',
        callback: action => {
          let display = document.getElementById('display')
          if (this.client) {
            // display.removeChild(this.client.getDisplay().getElement())
            display.innerHTML = ''
          }
        }
      })
    },

    sendClientClipboard(data) {
      if (!this.client) {
        return
      }
      let writer
      // Create stream with proper mimetype
      const stream = this.client.createClipboardStream(data.type)

      // Send data as a string if it is stored as a string
      if (typeof data.data === 'string') {
        writer = new Guacamole.StringWriter(stream)
        writer.sendText(data.data)
        writer.sendEnd()
      } else {
        // Write File/Blob asynchronously
        writer = new Guacamole.BlobWriter(stream)
        writer.oncomplete = function clipboardSent() {
          writer.sendEnd()
        }

        // Begin sending data
        writer.sendBlob(data.data)
      }
      console.log('send: ', data)
    },

    receiveClientClipboard(stream, mimetype) {
      console.log('recv: ', stream, mimetype)
      let reader
      // If the received data is text, read it as a simple string
      if (/^text\//.exec(mimetype)) {

        reader = new Guacamole.StringReader(stream)

        // Assemble received data into a single string
        let data = ''
        reader.ontext = function textReceived(text) {
          data += text
        }

        // Set clipboard contents once stream is finished
        reader.onend = async () => {
          this.clipboardText = data
          if (navigator.clipboard) {
            await navigator.clipboard.writeText(data)
          }
        }
      }

      // Otherwise read the clipboard data as a Blob
      else {
        reader = new Guacamole.BlobReader(stream, mimetype)
        reader.onprogress = function blobReceived(text) {
          console.log('blobReceived: ', text)
        }
        reader.onend = function end() {
          this.clipboardText = reader.getBlob()
        }
      }

    },

    oncursor(canvas, x, y) {
      this.localCursor = true
    },

    setLocalClipboard(data) {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(data)
      }
    },

    handleMouseState(mouseState) {

      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display)
        return

      // Send mouse state, show cursor if necessary
      this.display.showCursor(!this.localCursor)
      this.client.sendMouseState(mouseState, true)
    },

    onmousedown(mouseState) {
      document.body.focus()
      this.handleMouseState(mouseState)
      this.isMenuCollapse = true
    },

    onmouseout(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },

    onCloseDrawer() {
      console.log('onCloseDrawer', this.sink)
      this.sink.focus()
    },

    onWindowResize() {
      // 监听 window display的变化
      let pixel_density = window.devicePixelRatio || 1
      let optimal_width = window.innerWidth * pixel_density
      let optimal_height = window.innerHeight * pixel_density
      const width = optimal_width - 64
      const height = optimal_height
      if (this.client !== null) {
        const display = this.client.getDisplay()
        let displayHeight = display.getHeight() * pixel_density
        let displayWidth = display.getWidth() * pixel_density
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
      console.log('onWindowFocus   ')
      if (navigator.clipboard && navigator.clipboard.readText && this.clientState === 'Connected') {
        navigator.clipboard.readText().then((text) => {
          this.clipboardText = text
          this.sendClientClipboard({
            'data': text,
            'type': 'text/plain'
          })
        })
      }
    },

    fileSystemReceived(object, name) {
      console.log('fileSystemReceived ', object, name)
      this.currentFilesystem.object = object
      this.currentFilesystem.name = name
    },

    clientFileReceived(stream, mimetype, filename) {
      console.log('clientFileReceived, ', this.tunnel.uuid, stream, mimetype, filename)
      // Build download URL
      let url = BaseURL + this.apiPrefix
          + '/tunnels/' + encodeURIComponent(this.tunnel.uuid)
          + '/streams/' + encodeURIComponent(stream.index)
          + '/' + encodeURIComponent(sanitizeFilename(filename))

      // Create temporary hidden iframe to facilitate download
      let iframe = document.createElement('iframe')
      iframe.style.position = 'fixed'
      iframe.style.border = 'none'
      iframe.style.width = '1px'
      iframe.style.height = '1px'
      iframe.style.left = '-1px'
      iframe.style.top = '-1px'

      // The iframe MUST be part of the DOM for the download to occur
      document.body.appendChild(iframe)

      // Automatically remove iframe from DOM when download completes, if
      // browser supports tracking of iframe downloads via the "load" event
      iframe.onload = function downloadComplete() {
        document.body.removeChild(iframe)
      }

      // Acknowledge (and ignore) any received blobs
      stream.onblob = function acknowledgeData() {
        stream.sendAck('OK', Guacamole.Status.Code.SUCCESS)
      }

      // Automatically remove iframe from DOM a few seconds after the stream
      // ends, in the browser does NOT fire the "load" event for downloads
      stream.onend = function downloadComplete() {
        window.setTimeout(function cleanupIframe() {
          if (iframe.parentElement) {
            document.body.removeChild(iframe)
          }
        }, 500)
        this.loading = false
        this.loadingText = ''
      }.bind(this)
      this.loading = true
      this.loadingText = 'downloading file'
      // Begin download
      iframe.src = url
      console.log(url)

    },

    onsync: function(timestamp) {
      // console.log('onsync==> ', timestamp)
    },
    filedragenter: function(e) {
      e.stopPropagation()
      e.preventDefault()
    },
    filedragover: function(e) {
      e.stopPropagation()
      e.preventDefault()
    },
    filedrop: function(e) {
      e.stopPropagation()
      e.preventDefault()
      const dt = e.dataTransfer
      const files = dt.files
      this.handleFiles(files[0])
    },
    handleFiles: function(file, object, streamName) {
      let client = this.client
      let tunnel = this.tunnel
      let stream
      if (!object) {
        stream = client.createFileStream(file.type, file.name)
      } else {
        stream = object.createOutputStream(file.type, streamName)
      }
      let apiPrefix = this.apiPrefix
      return new Promise(function(resolve, reject) {
        // Upload file once stream is acknowledged
        stream.onack = function beginUpload(status) {

          // Notify of any errors from the Guacamole server
          if (status.isError()) {
            console.log(status.code, status)
            reject(status)
            return
          }
          let uploadToStream = function uploadStream(tunnel, stream, file,
                                                     progressCallback) {
            // Build upload URL
            let url = BaseURL + apiPrefix
                + '/tunnels/' + encodeURIComponent(tunnel)
                + '/streams/' + encodeURIComponent(stream.index)
                + '/' + encodeURIComponent(sanitizeFilename(file.name))
            let xhr = new XMLHttpRequest()
            // Invoke provided callback if upload tracking is supported
            if (progressCallback && xhr.upload) {
              xhr.upload.addEventListener('progress', function updateProgress(e) {
                progressCallback(e)
              })
            }
            // Resolve/reject promise once upload has stopped
            xhr.onreadystatechange = function uploadStatusChanged() {

              // Ignore state changes prior to completion
              if (xhr.readyState !== 4)
                return

              // Resolve if HTTP status code indicates success
              if (xhr.status >= 200 && xhr.status < 300) {
                console.log('success upload ')
                resolve()
              }
              // Parse and reject with resulting JSON error
              else if (xhr.getResponseHeader('Content-Type') === 'application/json')
                console.log('failed upload ', xhr.responseText)
              // Warn of lack of permission of a proxy rejects the upload
              else if (xhr.status >= 400 && xhr.status < 500) {
                console.log('failed upload ', xhr.status)
                reject(xhr.status)
              }
              // Assume internal error for all other cases
              else {
                console.log('failed upload ', xhr.status)
                reject(xhr.status)
              }
            }
            // Perform upload
            xhr.open('POST', url, true)
            const fd = new FormData()
            fd.append('file', file)
            xhr.send(fd)
          }
          // Begin upload
          uploadToStream(tunnel.uuid, stream, file, function uploadContinuing(event) {
            console.log('process upload ', event)
          })

          // Ignore all further acks
          stream.onack = null
        }
      })
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
      let dropbox = document.getElementById('display')
      dropbox.addEventListener('dragenter', this.filedragenter, false)
      dropbox.addEventListener('dragover', this.filedragover, false)
      dropbox.addEventListener('drop', this.filedrop, false)

      var display = document.getElementById('display')
      var tunnel = new Guacamole.WebSocketTunnel(wsURL)
      var client = new Guacamole.Client(tunnel)
      tunnel.onerror = function tunnelError(status) {
        console.log('tunnelError ', status)
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
      client.onclipboard = this.receiveClientClipboard
      client.onfilesystem = this.fileSystemReceived
      client.onfile = this.clientFileReceived
      client.onsync = this.onsync
      // Handle any received files
      client.connect(connectionParams)

      window.onunload = function() {
        client.disconnect()
      }
      var mouse = new Guacamole.Mouse(client.getDisplay().getElement())
      // Ensure focus is regained via mousedown before forwarding event
      mouse.onmousedown = this.onmousedown

      mouse.onmouseup = mouse.onmousemove = this.handleMouseState
      // Hide software cursor when mouse leaves display
      mouse.onmouseout = function() {
        if (!client.getDisplay()) return
        client.getDisplay().showCursor(false)
      }
      client.getDisplay().oncursor = this.oncursor
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

<style scoped>
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