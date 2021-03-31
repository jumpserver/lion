<template>
  <el-row>
    <i v-if="fileItem.type==='NORMAL'" class="el-icon-document"></i>
    <i v-else class="el-icon-folder"></i>
    <span @click="clickFile"> {{ fileItem.name }}</span>
  </el-row>

</template>

<script>
import {FileType} from '../utils/common'
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
      default: null,
    }
  },
  methods: {
    clickFile() {
      if (this.fileItem.type === FileType.NORMAL) {
        this.$emit('DownLoadFile', this.fileItem)
      } else {
        this.$emit('ChangeFolder', this.fileItem)
      }
    }
  }
}
</script>

<style scoped>

</style>