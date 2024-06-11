<template>
  <Connection :ws-url="wsUrl" :params="params" />
</template>

<script>
import Connection from './Connection'

export default {
  name: 'GuacamoleMonitor',
  components: {
    Connection
  },
  data() {
    return {
      readonly: false,
      wsUrl: '/lion/ws/monitor/',
      params: {
        'type': 'monitor',
        'SESSION_ID': this.$route.query.id
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
</style>
