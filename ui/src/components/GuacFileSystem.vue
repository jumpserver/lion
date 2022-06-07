<template>
  <div>
    <el-drawer
      direction="ltr"
      :title="$t('Files')"
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
        :on-success="refresh"
      >
        <el-button slot="trigger" size="small" type="primary">{{ $t('UploadFile') }}</el-button>
        <el-button size="small" type="default" style="margin-left: 10px" @click="clearFileList">{{ $t('ClearDone') }}</el-button>
      </el-upload>
      <div style="padding: 20px" class="fileZone">
        <el-row :gutter="20" class="currentFolder">
          <el-col :span="6">
            <span @click="changeParentFolder">{{ currentFolder.streamName }} </span>
          </el-col>
          <el-col :span="6" :offset="10">
            <i class="el-icon-refresh" style="padding-left: 20px;" @click="refresh" />
          </el-col>
        </el-row>
        <div class="fileList">
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
import { ErrorStatusCodes } from '@/utils/status'
import { getLanguage } from '@/i18n'

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
      fileList: [],
      currentFilesystem: {
        object: null,
        name: ''
      }
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
  },
  destroyed: function() {
    this.$log.debug('destroyed GuacFileSystem')
  },
  methods: {
    clearFileList() {
      this.fileList.splice(0, this.fileList.length)
    },
    fileSystemReceived(object, name) {
      this.$log.debug('fileSystemReceived ', object, name)
      this.currentFilesystem.object = object
      this.currentFilesystem.name = name

      this.updateDirectory(this.currentFolder).then(files => {
        this.$log.debug(files)
        this.files = files
      })
    },
    updateShow(value) {
      this.$log.debug('Update show to: ', value)
      this.$emit('update:show', value)
    },
    onCloseDrawer() {
      this.$emit('closeDrawer')
    },
    onDownloadFile(stream, mimetype, filename) {
      this.$log.debug('On download file')
      this.clientFileReceived(stream, mimetype, filename)
    },
    clientFileReceived(stream, mimetype, filename) {
      this.$log.debug('clientFileReceived, ', this.tunnel.uuid, stream, mimetype, filename)
      // Build download URL
      const url = BaseAPIURL +
          '/tunnels/' + encodeURIComponent(this.tunnel.uuid) +
          '/streams/' + encodeURIComponent(stream.index) +
          '/' + encodeURIComponent(sanitizeFilename(filename))

      // Create temporary hidden iframe to facilitate download
      const iframe = document.createElement('iframe')
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
        this.loadingText = ''
      }.bind(this)
      this.loadingText = 'downloading file'
      // Begin download
      iframe.src = url
      this.$log.debug(url)
    },

    fileDrop: function(e) {
      e.stopPropagation()
      e.preventDefault()
      const dt = e.dataTransfer
      const files = dt.files
      this.handleFiles(files[0]).then(() => {
        this.$message(files[0].name + ' ' + this.$t('UploadSuccess'))
      }).catch(status => {
        let msg = status.message
        if (getLanguage() === 'cn') {
          msg = this.$t(ErrorStatusCodes[status.code]) || status.message
        }
        this.$warning(msg)
      })
    },
    handleFiles: function(file, object, streamName, progressCallback) {
      const client = this.client
      const tunnel = this.tunnel
      const vm = this
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
            vm.$log.debug('Upload error', status.code, status)
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
            xhr.onreadystatechange = () => {
              // Ignore state changes prior to completion
              if (xhr.readyState !== 4) { return }

              // Resolve if HTTP status code indicates success
              if (xhr.status >= 200 && xhr.status < 300) {
                vm.$log.debug('Upload load success')
                resolve()
                // eslint-disable-next-line brace-style
              }
              // Parse and reject with resulting JSON error
              // eslint-disable-next-line brace-style
              else if (xhr.getResponseHeader('Content-Type') === 'application/json') {
                vm.$log.debug('failed upload ', xhr.responseText)
                // eslint-disable-next-line brace-style
              }
              // Warn of lack of permission of a proxy rejects the upload
              else if (xhr.status >= 400 && xhr.status < 500) {
                vm.$log.debug('Upload failed: ', xhr.status)
                reject(xhr.status)
                // eslint-disable-next-line brace-style
              }
              // Assume internal error for all other cases
              else {
                vm.$log.debug('Upload failed: ', xhr.status)
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
      this.$log.debug('uploadFile: ', fileObj)
      const file = fileObj.file
      let streamName
      if (this.currentFolder) {
        streamName = this.currentFolder.streamName + '/' + file.name
      }
      this.$log.debug('File is: ', file)
      const onprogress = function progress(e) {
        if (e.total > 0) {
          e.percent = e.loaded / e.total * 50
        }
        fileObj.onProgress(e)
      }
      this.handleFiles(file, this.currentFilesystem.object, streamName, onprogress).then((xhr) => {
        fileObj.onSuccess('Ok')
        this.refresh()
        this.$message(file.name + ' ' + this.$t('UploadSuccess'))
      }).catch(err => {
        fileObj.onError(err)
        this.$log.debug('Upload error: ', err)
        let msg = err.message
        if (getLanguage() === 'cn') {
          msg = this.$t(ErrorStatusCodes[err.code]) || err.message
        }
        this.$warning(msg)
      })
    },
    downloadFile(fileItem) {
      this.$log.debug('Download file: ', fileItem)
      const path = fileItem.streamName
      const downloadStreamReceived = function downloadStreamReceived(stream, mimetype) {
        // Parse filename from string
        const filename = path.match(/(.*[\\/])?(.*)/)[2]
        // Start download
        this.onDownloadFile(stream, mimetype, filename)
      }.bind(this)
      this.guacObject.requestInputStream(path, downloadStreamReceived)
    },

    changeFolder(fileItem) {
      // this.updateDirectory(fileItem)
      this.$log.debug('ChangeFolder ', fileItem)
      this.updateDirectory(fileItem).then(files => {
        this.files = files
        this.onChangeFolder(fileItem)
        // this.$emit('ChangeFolder', fileItem)
      })
    },
    changeParentFolder() {
      if (this.currentFolder.parent === null) {
        this.$log.debug('没有parent目录了')
        return
      }
      this.$log.debug('切换到parent目录了', this.currentFolder)
      this.updateDirectory(this.currentFolder.parent).then(files => {
        this.files = files
        this.changeFolder(this.currentFolder.parent)
      })
    },
    updateDirectory(file) {
      if (!this.guacObject) {
        return new Promise((resolve, reject) => {
          reject('No guacObject')
        })
      }
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

<style scoped>
.fileUploaderDraw ::v-deep .el-drawer__header {
  line-height: 30px;
  margin-bottom: 20px;
}

.upload-file {
  padding: 20px;
  border: 1px solid #ebebeb;
}

.upload-file ::v-deep .el-upload-list__item .el-icon-upload-success {
  color: #67C23A !important;
}

.fileZone {
  color: #409eff;
  font-size: 14px;
  line-height: 25px;
}

.el-icon-refresh:hover {
  color: #10355A;
}

.fileList {
  overflow: auto;
  padding-left: 10px;
  height: calc(100vh - 200px);
}

.currentFolder {
  cursor: pointer;
  font-weight: bold;
  padding-bottom: 20px
}
</style>
