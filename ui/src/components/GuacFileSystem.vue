<template>
  <div>
    <el-drawer
      direction="ltr"
      title="文件管理"
      :visible="show"
      class="fileUploaderDraw"
      @update:visible="updateShow"
      @close="onCloseDrawer"
    >
      <el-upload
        ref="upload"
        class="upload-file"
        action=""
        :multiple="true"
        :file-list="fileList"
        :auto-upload="true"
        :http-request="uploadFile"
      >
        <el-button slot="trigger" size="small" type="primary">上传文件</el-button>
<!--        <div slot="tip" class="el-upload__tip">只能上传jpg/png文件，且不超过500kb</div>-->
      </el-upload>
      <div style="padding: 20px" class="fileList">
        <el-row :gutter="20" style="padding-bottom: 20px">
          <el-col :span="6">
            <span @click="changeParentFolder">{{ currentFolder.streamName }} </span>
          </el-col>
          <el-col :span="6" :offset="10">
            <i class="el-icon-refresh" style="padding-left: 20px;" @click="refresh"></i>
          </el-col>
        </el-row>
        <div style="padding-left: 10px">
          <GuacFile
            v-for="(item ,index) in sortedFiles"
            :key="index"
            :file-item="item"
            @downloadFile="downloadFile"
            @changeFolder="changeFolder"
          />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { BaseAPIURL, FileType, isDirectory, sanitizeFilename } from '@/utils/common'
import GuacFile from './GuacFile'

export default {
  name: 'GuacFileSystem',
  components: {
    GuacFile
  },
  props: {
    client: {
      type: Object,
      required: true
    },
    tunnel: {
      type: Object,
      required: true
    },
    show: {
      type: Boolean,
      default: false
    },
    currentFilesystem: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      files: {},
      currentFolder: {
        mimetype: Guacamole.Object.STREAM_INDEX_MIMETYPE,
        streamName: Guacamole.Object.ROOT_STREAM,
        type: 'DIRECTORY',
        files: {},
        parent: null
      },
      fileList: []
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
      return ret
    },
    guacObject() {
      return this.currentFilesystem.object
    }
  },
  mounted: function() {
    console.log('mounted GuacFileSystem ', this.currentFolder)
    console.log('mounted guacObj', this.guacObject)
    this.updateDirectory(this.currentFolder).then(files => {
      console.log(files)
      this.files = files
    })
  },
  destroyed: function() {
    console.log('destroyed GuacFileSystem')
  },
  methods: {
    updateShow(value) {
      console.log('Update show to: ', value)
      this.$emit('update:show', value)
    },
    onCloseDrawer() {
      this.$emit('closeDrawer')
    },
    handleFiles: function(file, object, streamName, progressCallback) {
      const client = this.client
      const tunnel = this.tunnel
      let stream
      if (!object) {
        stream = client.createFileStream(file.type, file.name)
      } else {
        stream = object.createOutputStream(file.type, streamName)
      }
      return new Promise(function(resolve, reject) {
        // Upload file once stream is acknowledged
        stream.onack = function beginUpload(status) {
          // Notify of any errors from the Guacamole server
          if (status.isError()) {
            console.log('Upload error', status.code, status)
            reject(status)
            return
          }
          const uploadToStream = function uploadStream(tunnel, stream, file, progressCallback) {
            // Build upload URL
            const url = BaseAPIURL +
                '/tunnels/' + encodeURIComponent(tunnel) +
                '/streams/' + encodeURIComponent(stream.index) +
                '/' + encodeURIComponent(sanitizeFilename(file.name))
            const xhr = new XMLHttpRequest()
            xhr.withCredentials = true
            // Invoke provided callback if upload tracking is supported
            if (progressCallback && xhr.upload) {
              xhr.upload.addEventListener('progress', function updateProgress(e) {
                progressCallback(e)
              })
            }
            // Resolve/reject promise once upload has stopped
            xhr.onreadystatechange = function uploadStatusChanged() {
              // Ignore state changes prior to completion
              if (xhr.readyState !== 4) { return }

              // Resolve if HTTP status code indicates success
              if (xhr.status >= 200 && xhr.status < 300) {
                console.log('Upload load success')
                resolve()
                // eslint-disable-next-line brace-style
              }
              // Parse and reject with resulting JSON error
              // eslint-disable-next-line brace-style
              else if (xhr.getResponseHeader('Content-Type') === 'application/json') {
                console.log('failed upload ', xhr.responseText)
                // eslint-disable-next-line brace-style
              }
              // Warn of lack of permission of a proxy rejects the upload
              else if (xhr.status >= 400 && xhr.status < 500) {
                console.log('Upload failed: ', xhr.status)
                reject(xhr.status)
                // eslint-disable-next-line brace-style
              }
              // Assume internal error for all other cases
              else {
                console.log('Upload failed: ', xhr.status)
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
          uploadToStream(tunnel.uuid, stream, file, progressCallback)

          // Ignore all further acks
          stream.onack = null
        }
      })
    },
    onChangeFolder(fileItem) {
      this.currentFolder = fileItem
    },
    uploadFile(fileObj) {
      console.log('uploadFile: ', fileObj)
      const file = fileObj.file
      let streamName
      if (this.currentFolder) {
        streamName = this.currentFolder.streamName + '/' + file.name
      }
      console.log('File is: ', file)
      const onprogress = function progress(e) {
        if (e.total > 0) {
          e.percent = e.loaded / e.total * 100
        }
        fileObj.onProgress(e)
      }
      this.handleFiles(file, this.currentFilesystem.object, streamName, onprogress).then((xhr) => {
        fileObj.onSuccess('Ok')
        this.refresh()
      }).catch(err => {
        fileObj.onError(err)
        console.log('Upload error: ', err)
      })
    },
    downloadFile(fileItem) {
      console.log('Down load file: ', fileItem)
      const path = fileItem.streamName
      const downloadStreamReceived = function downloadStreamReceived(stream, mimetype) {
        // Parse filename from string
        const filename = path.match(/(.*[\\/])?(.*)/)[2]
        // Start download
        this.$emit('downloadReceived', stream, mimetype, filename)
      }.bind(this)
      this.guacObject.requestInputStream(path, downloadStreamReceived)
    },

    changeFolder(fileItem) {
      // this.updateDirectory(fileItem)
      console.log('ChangeFolder ', fileItem)
      this.updateDirectory(fileItem).then(files => {
        this.files = files
        this.onChangeFolder(fileItem)
        // this.$emit('ChangeFolder', fileItem)
      })
    },
    changeParentFolder() {
      if (this.currentFolder.parent === null) {
        console.log('没有parent目录了')
        return
      }
      console.log('切换到parent目录了', this.currentFolder)
      this.updateDirectory(this.currentFolder.parent).then(files => {
        this.files = files
        this.changeFolder(this.currentFolder.parent)
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
      this.files = []
      this.updateDirectory(this.currentFolder).then(files => {
        this.files = files
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.fileUploaderDraw {
  /deep/ .el-drawer__header {
    line-height: 30px;
    margin-bottom: 20px;
  }
}

.upload-file {
  padding: 20px;
  border: 1px solid #ebebeb;

  /deep/ .el-upload-list__item .el-icon-upload-success {
    color: #67C23A !important;
  }
}

.fileList {
  color: #409eff;
  font-size: 14px;
  cursor: pointer;
  line-height: 20px;
}

.el-icon-refresh:hover {
  color: #10355A;
}
</style>
