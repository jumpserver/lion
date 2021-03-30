<template>
  <el-container>
    <el-menu default-active="1" :collapse="isCollapse"
             @mouseleave.native="isCollapse = true"
             @mouseover.native="isCollapse = false"
    >

      <el-submenu index="1">
        <template slot="title">
          <i class="el-icon-position"></i>
          <span>快捷键</span>
        </template>
        <el-menu-item index="1-1">选项1</el-menu-item>
        <el-menu-item index="1-2">选项2</el-menu-item>
      </el-submenu>

      <el-menu-item><i class="el-icon-document-copy"></i>
        <span @click="toggleClipboard">剪切板</span>
        <el-drawer direction="ltr" :visible.sync="clipboardDrawer">
          <el-row>
            <el-col :span="12" :offset="8">
              <div class="grid-content bg-purple">
                <GuacClipboard v-bind:value="clipboardText" v-on:ClipboardChange="ClipboardChange"/>
              </div>
            </el-col>
          </el-row>
        </el-drawer>
      </el-menu-item>
      <el-menu-item><i class="el-icon-folder"></i>
        <span @click="toggleFile">文件管理</span>
        <el-drawer direction="ltr" :visible.sync="fileDrawer">
          <GuacFileSystem
              v-bind:guac-object="currentFilesystem.object"
              v-bind:name="currentFilesystem.name"
          />
        </el-drawer>
      </el-menu-item>

    </el-menu>
    <el-main>
      <el-row v-loading="loading" :element-loading-text="clientState" element-loading-background="rgba(0, 0, 0, 0.8">
        <div v-bind:style="divStyle" id="display"></div>
      </el-row>
    </el-main>
  </el-container>

</template>

<script>
import Guacamole from 'guacamole-common-js'
import { GetSupportedMimetypes } from '../utils/image'
import { sanitizeFilename } from '../utils/common'
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
      isCollapse: true,
      clipboardDrawer: false,
      fileDrawer: false,
      tunnelState: '',
      loading: true,
      clientState: '连中。。。',
      localCursor: false,
      client: null,
      tunnel: null,
      displayWidth: 0,
      displayHeight: 0,
      clipboardText: 'test',
      clipboardData: {
        type: 'text/plain',
        data: '',
      },
      currentFilesystem: {
        object: {},
        name: '',
        root: {
          mimetype: Guacamole.Object.STREAM_INDEX_MIMETYPE,
          streamName: Guacamole.Object.ROOT_STREAM,
          type: 'DIRECTORY',
          files: {},
        },
      },
      files: {}
    }
  },
  computed: {
    divStyle: function() {
      return {
        width: this.displayWidth + 'px',
        height: this.displayHeight + 'px'
      }
    }
  },
  mounted: function() {
    window.addEventListener('resize', this.onWindowResize)
    window.onfocus = this.onWindowFocus
    this.getConnectString('0000-0000-00').then(connectionParams => {
      console.log(connectionParams)
      console.log(this.displayWidth, this.displayHeight)
      this.createGuacamole(connectionParams)
    })
  },
  methods: {
    ClipboardChange(data) {
      console.log('ClipboardChange emit ', data)
      this.clipboardText = data
    },
    toggleClipboard() {
      this.clipboardDrawer = !this.clipboardDrawer
    },
    toggleFile() {
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
          break

      }
    },

    clientOnErr(stats) {
      this.client.disconnect()
      console.log(stats)
    },

    sendClientClipboard(data) {
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

      this.clipboardText = data.data

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
    },

    onmouseout(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },

    onWindowResize() {
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
      this.displayWidth = width
      this.displayHeight = height
      console.log(width, height)
    },

    onWindowFocus() {
      console.log('onWindowFocus   ')
      if (navigator.clipboard && navigator.clipboard.readText && this.clientState === 'Connected') {
        navigator.clipboard.readText().then((text) => {
          if (this.clipboardText === text) {
            console.log('内容一样，可以不发送')
            return
          }
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

      // this.currentFilesystem.object.requestInputStream(this.currentFilesystem.root.streamName, this.handleStream)
    },

    handleStream(stream, mimetype) {
      // Ignore stream if mimetype is wrong
      if (mimetype !== Guacamole.Object.STREAM_INDEX_MIMETYPE) {
        stream.sendAck('Unexpected mimetype', Guacamole.Status.Code.UNSUPPORTED)
        return
      }

      // Signal server that data is ready to be received
      stream.sendAck('Ready', Guacamole.Status.Code.SUCCESS)

      // Read stream as JSON
      var reader = new Guacamole.JSONReader(stream)

      // Acknowledge received JSON blobs
      reader.onprogress = function onprogress() {
        stream.sendAck('Received', Guacamole.Status.Code.SUCCESS)
      }

      // Reset contents of directory
      var files = {}
      reader.onend = function jsonReady() {

        // Determine the expected filename prefix of each stream
        var expectedPrefix = '/'
        if (expectedPrefix.charAt(expectedPrefix.length - 1) !== '/')
          expectedPrefix += '/'

        // For each received stream name
        var mimetypes = reader.getJSON()
        console.log(mimetypes)
        for (var name in mimetypes) {
          console.log(name)
          // Assert prefix is correct
          if (name.substring(0, expectedPrefix.length) !== expectedPrefix)
            continue

          // Extract filename from stream name
          var filename = name.substring(expectedPrefix.length)

          // Deduce type from mimetype
          var type = 'NORMAL'
          if (mimetypes[name] === Guacamole.Object.STREAM_INDEX_MIMETYPE)
            type = 'DIRECTORY'

          // Add file entry
          console.log(this.files)
          files[filename] = {
            mimetype: mimetypes[name],
            streamName: name,
            type: type,
            parent: '/',
            name: filename
          }

        }
      }


    },

    clientFileReceived(stream, mimetype, filename) {
      console.log('clientFileReceived, ', this.tunnel.uuid, stream, mimetype, filename)

      //  tunnelService.downloadStream(tunnel.uuid, stream, mimetype, filename);
      // Work-around for IE missing window.location.origin
      if (!window.location.origin)
        var streamOrigin = window.location.protocol + '//' + window.location.hostname + (window.location.port ? (':' + window.location.port) : '')
      else
        var streamOrigin = window.location.origin

      // Build download URL
      var url = streamOrigin + '/guacamole'
          + '/api/tunnels/' + encodeURIComponent(this.tunnel.uuid)
          + '/streams/' + encodeURIComponent(stream.index)
          + '/' + encodeURIComponent(sanitizeFilename(filename))

      // Create temporary hidden iframe to facilitate download
      var iframe = document.createElement('iframe')
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
        console.log('on blob')
        stream.sendAck('OK', Guacamole.Status.Code.SUCCESS)
      }

      // Automatically remove iframe from DOM a few seconds after the stream
      // ends, in the browser does NOT fire the "load" event for downloads
      stream.onend = function downloadComplete() {
        window.setTimeout(function cleanupIframe() {
          if (iframe.parentElement) {
            document.body.removeChild(iframe)
          }
        }, 5000)
      }

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

      this.handleFiles(files)
    },
    handleFiles: function(files) {
      console.log(files)
      let file = files[0]
      console.log(file.type, file.name)
      let client = this.client
      let tunnel = this.tunnel
      // let stream;
      let stream = client.createFileStream(file.type, file.name)
      // Upload file once stream is acknowledged
      stream.onack = function beginUpload(status) {

        // Notify of any errors from the Guacamole server
        if (status.isError()) {
          console.log(status.code, status)
          return
        }

        let uploadToStream = function uploadStream(tunnel, stream, file,
                                                   progressCallback) {
          let streamOrigin
          if (!window.location.origin)
            streamOrigin = window.location.protocol + '//' + window.location.hostname + (window.location.port ? (':' + window.location.port) : '')
          else
            streamOrigin = window.location.origin
          // Build upload URL
          let url = streamOrigin + window.location.pathname
              + '/api/tunnels/' + encodeURIComponent(tunnel)
              + '/streams/' + encodeURIComponent(stream.index)
              + '/' + encodeURIComponent(sanitizeFilename(file.name))
          let xhr = new XMLHttpRequest()
          // Invoke provided callback if upload tracking is supported
          if (progressCallback && xhr.upload) {
            xhr.upload.addEventListener('progress', function updateProgress(e) {
              progressCallback(e.loaded)
            })
          }

          // Resolve/reject promise once upload has stopped
          xhr.onreadystatechange = function uploadStatusChanged() {

            // Ignore state changes prior to completion
            if (xhr.readyState !== 4)
              return

            // Resolve if HTTP status code indicates success
            if (xhr.status >= 200 && xhr.status < 300)
              console.log('success upload ')

            // Parse and reject with resulting JSON error
            else if (xhr.getResponseHeader('Content-Type') === 'application/json')
              console.log('failed upload ', xhr.responseText)
            // Warn of lack of permission of a proxy rejects the upload
            else if (xhr.status >= 400 && xhr.status < 500)
              console.log('failed upload ', xhr.status)

            // Assume internal error for all other cases
            else
              console.log('failed upload ', xhr.status)
          }

          // Perform upload
          xhr.open('POST', url, true)
          const fd = new FormData()
          fd.append('file', file)
          xhr.send(fd)
        }
        // Begin upload
        uploadToStream(tunnel.uuid, stream, file, function uploadContinuing(length) {
          console.log('process ', length)
        })

        // Ignore all further acks
        stream.onack = null

      }
    },

    createGuacamole(connectionParams) {

      let dropbox = document.getElementById('display')
      console.log(dropbox)
      dropbox.addEventListener('dragenter', this.filedragenter, false)
      dropbox.addEventListener('dragover', this.filedragover, false)
      dropbox.addEventListener('drop', this.filedrop, false)

      var display = document.getElementById('display')
      var tunnel = new Guacamole.WebSocketTunnel('/guacamole/ws')
      var client = new Guacamole.Client(tunnel)
      tunnel.onerror = function tunnelError(status) {
        console.log('tunnelError ', status)
      }
      tunnel.onuuid = function tunnelAssignedUUID(uuid) {
        console.log('tunnelAssignedUUID ', uuid)
        tunnel.uuid = uuid
      }

      tunnel.onstatechange = this.onTunnelStateChanged
      this.client = client
      this.tunnel = tunnel
      this.display = this.client.getDisplay()
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
      var keyboard = new Guacamole.Keyboard(sink.getElement())

      keyboard.onkeydown = function(keysym) {
        client.sendKeyEvent(1, keysym)
      }

      keyboard.onkeyup = function(keysym) {
        client.sendKeyEvent(0, keysym)
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

.el-dropdown-link {
  cursor: pointer;
  color: #409EFF;
}

.el-icon-arrow-down {
  font-size: 12px;
}
</style>