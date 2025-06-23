<template>
  <el-drawer direction="ltr" class="clipboardDrawer" :title="$t('Clipboard')" :visible="visible" @update:visible="updateVisible" @close="onCloseDrawer">
    <div class="grid-content bg-purple" style="width: 100%">
      <el-input v-model="clipboardText" type="textarea" class="clipboard" :rows="10" />
    </div>
  </el-drawer>
</template>

<script>
import Guacamole from 'guacamole-common-js'

export default {
  name: 'GuacClipboard',
  props: {
    client: {
      type: Object,
      required: true
    },
    tunnel: {
      type: Object,
      required: true
    },
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      clipboardText: ''
    }
  },
  watch: {
    clipboardText(newValue) {
      this.sendDataToRemoteClipboard(newValue)
      this.setLocalClipboard(newValue)
    }
  },
  methods: {
    sendClipboardToRemote() {
      if (navigator.clipboard && navigator.clipboard.readText) {
        navigator.clipboard.readText().then((text) => {
          if (text !== this.clipboardText) {
            this.sendDataToRemoteClipboard(text)
            this.clipboardText = text
          }
        })
      }
    },
    updateVisible(value) {
      this.$emit('update:visible', value)
    },
    sendDataToRemoteClipboard(value) {
      const data = {
        type: 'text/plain',
        data: value
      }
      if (!this.client) {
        return
      }
      this.clipboardText = value
      let writer
      // Create stream with proper mimetype
      const stream = this.client.createClipboardStream(data.type)

      // Send data as a string if it is stored as a string
      if (typeof data.data === 'string') {
        writer = new Guacamole.StringWriter(stream)
        writer.sendText(data.data)
        writer.sendEnd()
        this.$log.debug('send text: ', data)
      } else {
        // Write File/Blob asynchronously
        writer = new Guacamole.BlobWriter(stream)
        writer.oncomplete = function clipboardSent() {
          writer.sendEnd()
        }
        // Begin sending data
        this.$log.debug('Send blob: ', data)
        writer.sendBlob(data.data)
      }
    },
    receiveClientClipboard(stream, mimetype) {
      this.$log.debug('Recv clipboard: ', stream, mimetype)
      let reader
      // If the received data is text, read it as a simple string
      if (/^text\//.exec(mimetype)) {
        reader = new Guacamole.StringReader(stream)

        // Assemble received data into a single string
        let data = ''
        reader.ontext = (text) => {
          data += text
        }

        // Set clipboard contents once stream is finished
        reader.onend = async() => {
          console.log('clipboard received from remote: ', data)
          if (navigator.clipboard) {
            await navigator.clipboard.writeText(data)
          }
        }
        // eslint-disable-next-line brace-style
      }
      // Otherwise read the clipboard data as a Blob
      else {
        reader = new Guacamole.BlobReader(stream, mimetype)
        reader.onprogress = (text) => {
          this.$log.debug('blobReceived: ', text)
        }
        reader.onend = () => {
          this.clipboardText = reader.getBlob()
        }
      }
    },
    setLocalClipboard(data) {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(data)
      }
    },
    onCloseDrawer() {
      this.$emit('closeDrawer')
    }
  }
}
</script>

<style scoped>
.clipboard {
  position: relative;
  -moz-border-radius: 0.25em;
  -webkit-border-radius: 0.25em;
  border-radius: 0.25em;
  white-space: pre;
  font-size: 1em;
  overflow: auto;
  padding-left: 10px;
  height: 100%;
  width: calc(100% - 20px);

}

.clipboard div {
  margin: 0;
}

.clipboard .el-textarea__inner {
  background-color: #303133;
}

</style>
