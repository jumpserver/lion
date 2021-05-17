<template>
  <el-main>
    <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="rgba(0, 0, 0, 0.8">
      <div :style="divStyle">
        <div id="monitor" />
      </div>
    </el-row>
  </el-main>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import i18n from '@/i18n'
import { getMonitorConnectParams } from '../utils/common'
import { getSupportedMimetypes } from '../utils/image'
import { getSupportedGuacAudios } from '../utils/audios'
import { getSupportedGuacVideos } from '../utils/video'
import { getLanguage } from '../i18n'
import { ErrorStatusCodes } from '@/utils/status'

export default {
  name: 'GuacamoleMonitor',
  data() {
    return {
      displayWidth: 0,
      displayHeight: 0,
      loading: true,
      loadingText: i18n.t('Connecting') + ' ...'
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
    const result = getMonitorConnectParams()
    this.$log.debug(result)
    const sid = result['data']['session']
    this.getConnectString(sid).then(connectionParams => {
      this.connectGuacamole(connectionParams, result['ws'])
    }).catch(err => {
      this.$log.debug(err)
    })
  },
  methods: {

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

    displayResize(width, height) {
      // 监听guacamole display的变化
      this.displayWidth = width
      this.displayHeight = height
      window.resizeTo(width, height)
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
          this.loadingText = 'Connecting'
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
            // Create new audio stream, associating it wit
            // AudioRecorder
            var stream = client.createAudioStream(AUDIO_INPUT_MIMETYPE)
            var recorder = Guacamole.AudioRecorder.getInstance(stream, AUDIO_INPUT_MIMETYPE)

            // If creation of the AudioRecorder failed, simply end the stream
            // eslint-disable-next-line brace-style
            if (!recorder) { stream.sendEnd() }

            // Otherwise, ensure that another audio stream is created after this
            // audio stream is closed
            else {
              recorder.onclose = requestAudioStream.bind(this, client)
            }
            this.$log.debug(stream, recorder)
          }
          requestAudioStream(this.client)
          break

          // Update history when disconnecting
        case 4: // Disconnecting
        case 5: // Disconnected
          this.clientState = 'Disconnecting'
          this.$log.debug('clientState, Disconnected ')
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
    connectGuacamole(connectionParams, wsURL) {
      var display = document.getElementById('monitor')
      var tunnel = new Guacamole.WebSocketTunnel(wsURL)
      var client = new Guacamole.Client(tunnel)
      tunnel.onerror = function tunnelError(status) {
        this.$log.debug('tunnelError ', status)
        display.innerHTML = ''
      }
      tunnel.onuuid = function tunnelAssignedUUID(uuid) {
        this.$log.debug('tunnelAssignedUUID ', uuid)
        tunnel.uuid = uuid
      }
      tunnel.onstatechange = this.onTunnelStateChanged
      this.client = client
      this.tunnel = tunnel
      this.display = this.client.getDisplay()
      this.display.onresize = this.displayResize
      display.appendChild(client.getDisplay().getElement())
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
