<template>
  <div class="file-browser">
    <GuacFile v-for="(item ,index) in currentFolder.files" :key="index" v-bind:fileItem="item"
              v-on:DownLoadFile="DownLoadFile"
              v-on:ChangeFolder="ChangeFolder"/>
  </div>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { FileType } from '../utils/common'
import GuacFile from './GuacFile'

export default {
  name: 'GuacFileSystem',
  components: {
    GuacFile,
  },
  props: {
    guacObject: {
      type: Object,
      default: null
    },
    currentFolder: {
      type: Object,
      default: null
    }
  },
  mounted: function() {
    console.log('mounted GuacFileSystem ', this.currentFolder)
    this.updateDirectory(this.currentFolder)

  },
  destroyed: function() {
    console.log('destroyed GuacFileSystem')
  },
  methods: {
    DownLoadFile(fileItem) {
      let path = fileItem.streamName
      let downloadStreamReceived = function downloadStreamReceived(stream, mimetype) {
        // Parse filename from string
        var filename = path.match(/(.*[\\/])?(.*)/)[2]
        // Start download
        this.$emit('DownLoadReceived', stream, mimetype, filename)
      }.bind(this)
      this.guacObject.requestInputStream(path, downloadStreamReceived)
    },

    ChangeFolder(fileItem) {
      // this.updateDirectory(fileItem)
      console.log('ChangeFolder ', fileItem)
      this.$emit('ChangeFolder', fileItem)
    },
    ChangeParentFolder() {
      if (this.currentFolder.parent === null) {
        console.log('没有parent目录了')
        return
      }
      console.log('切换到parent目录了', this.currentFolder)
      this.$emit('ChangeParentFolder', this.currentFolder.parent)
    },
    createFile(template) {
      return {
        mimetype: template.mimetype,
        streamName: template.streamName,
        type: template.type,
        name: template.name,
        parent: template.parent,
        files: template.files || {},
      }
    },
    updateDirectory(file) {
      // Do not attempt to refresh the contents of directories
      if (file.mimetype !== Guacamole.Object.STREAM_INDEX_MIMETYPE)
        return

      // Request contents of given file
      this.guacObject.requestInputStream(file.streamName, function handleStream(stream, mimetype) {

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
        reader.onend = function jsonReady() {
          // Empty contents
          file.files = {}

          // Determine the expected filename prefix of each stream
          var expectedPrefix = file.streamName
          if (expectedPrefix.charAt(expectedPrefix.length - 1) !== '/')
            expectedPrefix += '/'

          // For each received stream name
          var mimetypes = reader.getJSON()
          for (var name in mimetypes) {

            // Assert prefix is correct
            if (name.substring(0, expectedPrefix.length) !== expectedPrefix)
              continue

            // Extract filename from stream name
            var filename = name.substring(expectedPrefix.length)

            // Deduce type from mimetype
            var type = FileType.NORMAL
            if (mimetypes[name] === Guacamole.Object.STREAM_INDEX_MIMETYPE)
              type = FileType.DIRECTORY

            // Add file entry
            file.files[filename] = {
              mimetype: mimetypes[name],
              streamName: name,
              type: type,
              parent: file,
              name: filename
            }
            console.log(file.files)

          }
        }

      })
    }
  }
}
</script>

<style scoped>
.file-browser .directory > .children {
  padding-left: 1em;
  display: none;
}

.file-browser .list-item .caption {
  white-space: nowrap;
  border: 1px solid transparent;
}

.file-browser .list-item.focused .caption {
  border: 1px dotted rgba(0, 0, 0, 0.5);
  background: rgba(204, 221, 170, 0.5);
}

</style>