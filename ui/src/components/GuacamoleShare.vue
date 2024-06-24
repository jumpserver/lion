<template>
  <div>
    <div v-if="!codeDialog">
      <Connection :readonly="readonly" :ws-url="wsUrl" :params="params" />
    </div>
    <el-dialog
      :title="$t('VerifyCode')"
      :visible.sync="codeDialog"
      :close-on-press-escape="false"
      :close-on-click-modal="false"
      :show-close="false"
      width="30%"
    >
      <el-form ref="form" class="code-form" @submit.native.prevent>
        <el-form-item>
          <el-input v-model="code" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button class="item-button" @click="submitCode">{{ $t('ConfirmBtn') }}</el-button>
      </div>
    </el-dialog>
  </div>

</template>

<script>
import Connection from './Connection'
import { getShareSession } from '@/api/session'

export default {
  name: 'GuacamoleShare',
  components: {
    Connection
  },
  data() {
    return {
      readonly: false,
      wsUrl: '/lion/ws/share/',
      codeDialog: true,
      code: '',
      sessionId: ''
    }
  },
  computed: {
    params() {
      return {
        'type': 'share',
        'SESSION_ID': this.sessionId,
        'SHARE_ID': this.$route.params.id
      }
    }
  },
  methods: {
    submitCode() {
      if (this.code === '') {
        this.$message(this.$t('InputVerifyCode'))
        return
      }
      const data = {
        code: this.code
      }
      getShareSession(this.$route.params.id, data).then(res => {
        this.$log.debug(res)
        const actionPerm = res['action_permission']
        if (actionPerm['value'] === 'readonly') {
          this.readonly = true
        }
        this.sessionId = res.session.id
        this.codeDialog = false
      }).catch(err => {
        this.$log.error(err)
        this.$message.error(err)
      })
    }
  }
}
</script>

<style scoped>
.el-container {
  margin: 0 auto;
}

.el-main {
  padding: 0;
}

.code-form >>> .el-input__inner:hover {
  border-color:  #303133;
}

.item-button >>> .el-button {
  background: #303133;
  color: #ffffff;
}

.item-button:hover {
  background: rgb(134, 133, 133);
  color: #303133;
  border-color: rgba(183, 172, 172, 0.1);
}

</style>
