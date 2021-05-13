<template>
  <el-drawer direction="ltr" title="剪切板" :visible="visible" @update:visible="updateVisible" @close="onCloseDrawer">
    <div class="grid-content bg-purple" style="width: 100%">
      <el-input v-model="value" type="textarea" class="clipboard" :rows="10" />
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
      value: ''
    }
  },
  watch: {
    value(newValue) {
      this.sendClipboardToRemote(newValue)
    }
  },
  methods: {
    updateVisible(value) {
      this.$emit('update:visible', value)
    },
    sendClipboardToRemote(value) {
      const data = {
        type: 'text/plain',
        data: value
      }
      if (!this.client) {
        return
      }
      this.clipboardText = data
      this.setLocalClipboard(data)
      let writer
      // Create stream with proper mimetype
      const stream = this.client.createClipboardStream(data.type)

      // Send data as a string if it is stored as a string
      console.log('Typeof data: ', typeof data.data)
      if (typeof data.data === 'string') {
        writer = new Guacamole.StringWriter(stream)
        writer.sendText(data.data)
        writer.oncomplete = () => {
          writer.sendEnd()
          console.log('send done: ', data)
        }
        console.log('send text: ', data)
      } else {
        // Write File/Blob asynchronously
        writer = new Guacamole.BlobWriter(stream)
        writer.oncomplete = function clipboardSent() {
          writer.sendEnd()
        }
        // Begin sending data
        console.log('Send blob: ', data)
        writer.sendBlob(data.data)
      }
    },
    receiveClientClipboard(stream, mimetype) {
      console.log('Recv clipboard: ', stream, mimetype)
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
        reader.onend = async() => {
          this.clipboardText = data
          if (navigator.clipboard) {
            await navigator.clipboard.writeText(data)
          }
        }
        // eslint-disable-next-line brace-style
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

</style>
