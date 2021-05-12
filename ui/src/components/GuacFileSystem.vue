<template>
  <div style="padding: 20px">
    <span>{{ guacObject.name }}</span>
    <el-row :gutter="20">
      <el-col :span="6"><span @click="ChangeParentFolder">{{ currentFolder.streamName }} </span></el-col>
      <el-col :span="6" :offset="10">
        <el-button size="small" type="primary" @click="clickFile">点击上传</el-button>
        <input ref="fileinput" type="file" hidden @change="uploadFile">
      </el-col>
    </el-row>
    <GuacFile
      v-for="(item ,index) in sortedFiles"
      :key="index"
      :file-item="item"
      @DownLoadFile="downLoadFile"
      @ChangeFolder="changeFolder"
    />
  </div>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { FileType, isDirectory } from '../utils/common'
import GuacFile from './GuacFile'

export default {
  name: 'GuacFileSystem',
  components: {
    GuacFile
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
  data() {
    return {
      files: {}
    }
  },
  computed: {
    sortedFiles: function() {
      const unsortedFiles = []
      for (const name in this.files) { unsortedFiles.push(this.files[name]) }
      const ret = unsortedFiles.sort(function fileComparator(a, b) {
        // Directories come before non-directories
        if (isDirectory(a) && !isDirectory(b)) { return -1 }

        // Non-directories come after directories
        if (!isDirectory(a) && isDirectory(b)) { return 1 }

        // All other combinations are sorted by name
        return a.name.localeCompare(b.name)
      })
      console.log(ret)
      return ret
    }
  },

  mounted: function() {
    console.log('mounted GuacFileSystem ', this.currentFolder)
    this.updateDirectory(this.currentFolder).then(files => {
      console.log(files)
      this.files = files
    })
  },

  destroyed: function() {
    console.log('destroyed GuacFileSystem')
  },

  methods: {
    clickFile() {
      this.$refs.fileinput.click()
    },
    uploadFile(event) {
      this.$emit('UploadFile', event.target.files)
    },
    downLoadFile(fileItem) {
      const path = fileItem.streamName
      const downloadStreamReceived = function downloadStreamReceived(stream, mimetype) {
        // Parse filename from string
        var filename = path.match(/(.*[\\/])?(.*)/)[2]
        // Start download
        this.$emit('DownLoadReceived', stream, mimetype, filename)
      }.bind(this)
      this.guacObject.requestInputStream(path, downloadStreamReceived)
    },

    changeFolder(fileItem) {
      // this.updateDirectory(fileItem)
      console.log('ChangeFolder ', fileItem)
      this.updateDirectory(fileItem).then(files => {
        this.files = files
        this.$emit('ChangeFolder', fileItem)
      })
    },
    ChangeParentFolder() {
      if (this.currentFolder.parent === null) {
        console.log('没有parent目录了')
        return
      }
      console.log('切换到parent目录了', this.currentFolder)
      this.updateDirectory(this.currentFolder.parent).then(files => {
        this.files = files
        this.$emit('ChangeFolder', this.currentFolder.parent)
      })
    },
    updateDirectory(file) {
      const guacObject = this.guacObject
      return new Promise(function(resolve, reject) {
        // Do not attempt to refresh the contents of directories
        if (file.mimetype !== Guacamole.Object.STREAM_INDEX_MIMETYPE) { return }
        // Request contents of given file
        guacObject.requestInputStream(file.streamName, function handleStream(stream, mimetype) {
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
            if (expectedPrefix.charAt(expectedPrefix.length - 1) !== '/') { expectedPrefix += '/' }

            // For each received stream name
            var mimetypes = reader.getJSON()
            for (var name in mimetypes) {
              // Assert prefix is correct
              if (name.substring(0, expectedPrefix.length) !== expectedPrefix) { continue }

              // Extract filename from stream name
              var filename = name.substring(expectedPrefix.length)

              // Deduce type from mimetype
              var type = FileType.NORMAL
              if (mimetypes[name] === Guacamole.Object.STREAM_INDEX_MIMETYPE) { type = FileType.DIRECTORY }

              // Add file entry
              file.files[filename] = {
                mimetype: mimetypes[name],
                streamName: name,
                type: type,
                parent: file,
                name: filename
              }
            }
            resolve(file.files)
          }
        })
      })
    },

    refresh() {
      this.updateDirectory(this.currentFolder).then(files => {
        this.files = files
      })
    }
  }
}
</script>

<style scoped>

</style>
