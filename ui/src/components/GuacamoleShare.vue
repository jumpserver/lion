<template>
  <div>
    <div v-if="!codeDialog">
      <Connection :readonly="readonly" :ws-url="wsUrl" :params="params" @jms-event="OnJmsEvent" />
    </div>
    <RightPanel ref="panel">
      <Settings :settings="settings" :title="$t('Settings')" />
    </RightPanel>
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
import Settings from '@/components/Settings.vue'
import RightPanel from '@/components/RightPanel.vue'

export default {
  name: 'GuacamoleShare',
  components: {
    Connection,
    RightPanel,
    Settings
  },
  data() {
    return {
      readonly: false,
      wsUrl: '/lion/ws/share/',
      codeDialog: true,
      code: '',
      sessionId: '',
      onlineUsersMap: {},
      share_id: '',
      recordId: ''
    }
  },
  computed: {
    params() {
      return {
        'type': 'share',
        'SESSION_ID': this.sessionId,
        'SHARE_ID': this.$route.params.id,
        'RECORD_ID': this.recordId
      }
    },
    settings() {
      return [
        {
          title: this.$t('User'),
          icon: 'el-icon-s-custom',
          disabled: () => Object.keys(this.onlineUsersMap).length > 1,
          content: Object.values(this.onlineUsersMap).map(item => {
            item.name = (this.share_id !== item.share_id) ? item.user : item.user + ' [' + this.$t('Self') + ']'
            item.faIcon = item.writable ? 'fa-solid fa-keyboard' : 'fa-solid fa-eye'
            item.iconTip = item.writable ? this.$t('Writable') : this.$t('ReadOnly')
            return item
          }).sort((a, b) => new Date(a.created) - new Date(b.created)),
          itemClick: () => {}
        }
      ]
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
        this.recordId = res.id
        this.sessionId = res.session.id
        this.codeDialog = false
      }).catch(err => {
        this.$log.error(err)
        this.$message.error(err)
      })
    },
    OnJmsEvent(event, data) {
      this.$log.debug(event, data)
      const dataObj = JSON.parse(data)
      switch (event) {
        case 'share_join': {
          const key = dataObj.share_id
          this.$set(this.onlineUsersMap, key, data)
          this.$log.debug(this.onlineUsersMap)
          if (dataObj.primary) {
            this.$log.debug('primary user 不提醒')
            break
          }
          const joinMsg = `${data.user} ${this.$t('JoinShare')}`
          this.$message(joinMsg)
          break
        }
        case 'share_exit': {
          const key = dataObj.share_id
          this.$delete(this.onlineUsersMap, key)
          const leaveMsg = `${data.user} ${this.$t('LeaveShare')}`
          this.$message(leaveMsg)
          break
        }
        case 'share_users': {
          this.userOptions = dataObj
          break
        }
      }
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
