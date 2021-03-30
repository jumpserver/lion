<template>
  <el-row>
    <i v-if="fileItem.type==='NORMAL'" class="el-icon-document"></i>
    <i v-else class="el-icon-folder"></i>
    <span @click="clickFile"> {{ fileItem.name }}</span>
  </el-row>

</template>

<script>
import { FileType } from '../utils/common'

export default {
  name: 'GuacFile',
  // props: {
  //   mimetype: {
  //     type: String,
  //     default: ''
  //   },
  //   streamName: {
  //     type: String,
  //     default: ''
  //   },
  //   type: {
  //     type: String,
  //     default: FileType.NORMAL
  //   },
  //   name: {
  //     type: String,
  //     default: ''
  //   },
  //   parent: {
  //     type: Object,
  //     default: null
  //   },
  //   files: {
  //     type: Object,
  //     default: () => ({}),
  //   }
  // },
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
      console.log('click clickFile',this.fileItem)
    }
  }
}
</script>

<style scoped>

</style>