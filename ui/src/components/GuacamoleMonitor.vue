<template>
  <Connection :ws-url="wsUrl" :params="params" :readonly="readonly" />
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
      readonly: true,
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
