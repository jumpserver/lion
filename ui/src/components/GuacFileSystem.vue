<template>
  <el-row class="file-browser">
    <GuacFile v-for="(item ,index) in currentDirectory.files" :key="index" v-bind="item"></GuacFile>
  </el-row>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import {FileType} from '../utils/common'
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
    name: {
      type: String,
      default: ''
    }
  },
  data() {
    const root = this.createFile({
      mimetype: Guacamole.Object.STREAM_INDEX_MIMETYPE,
      streamName: Guacamole.Object.ROOT_STREAM,
      type: FileType.DIRECTORY

    })
    return {
      root: this.createFile(root),
      currentDirectory: this.createFile(root),
      files: {},
    }
  },
  mounted: function() {
    this.updateDirectory(this.currentDirectory)
  },
  methods: {
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