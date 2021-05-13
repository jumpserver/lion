<template>
  <div class="fileItem">
    <i v-if="fileItem.type==='NORMAL'" class="el-icon-document" />
    <i v-else class="el-icon-folder" />
    <el-link class="filename" @click="clickFile"> {{ filename }} </el-link>
  </div>
</template>

<script>
import { FileType } from '@/utils/common'
/*
{
  mimetype:   String,
  streamName: String,
  type:       FileType.NORMAL,
  name:       String,
  parent:     FILE,
  files:      {},
}
 */
export default {
  name: 'GuacFile',
  props: {
    fileItem: {
      type: Object,
      default: null
    }
  },
  computed: {
    filename() {
      const filename = this.fileItem.name
      const filenameLen = filename.length
      if (filenameLen < 45) {
        return filename
      } else {
        return filename.slice(0, 30) + '...' + filename.slice(filenameLen - 15, filenameLen)
      }
    }
  },
  methods: {
    clickFile() {
      if (this.fileItem.type === FileType.NORMAL) {
        this.$emit('downloadFile', this.fileItem)
      } else {
        this.$emit('changeFolder', this.fileItem)
      }
    }
  }
}
</script>

<style scoped>
.fileItem {
  cursor: pointer;
}

.filename {
  padding-left: 6px;
}
</style>
