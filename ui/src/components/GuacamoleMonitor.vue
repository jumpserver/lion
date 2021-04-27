<template>
  <el-main>
    <el-row v-loading="loading" :element-loading-text="loadingText" element-loading-background="rgba(0, 0, 0, 0.8">
      <div :style="divStyle">
        <div id="monitor"/>
      </div>
    </el-row>
  </el-main>
</template>

<script>
export default {
  name: 'GuacamoleMonitor',
  data() {
    return {
      divStyle: {},
      loading: true,
      loadingText: '连接中。。',
    }
  },
  methods: {
    connectGuacamole(connectionParams, wsURL) {
      var display = document.getElementById('monitor')
      var tunnel = new Guacamole.WebSocketTunnel(wsURL)
      var client = new Guacamole.Client(tunnel)
      tunnel.onerror = function tunnelError(status) {
        console.log('tunnelError ', status)
      }
      tunnel.onuuid = function tunnelAssignedUUID(uuid) {
        console.log('tunnelAssignedUUID ', uuid)
        tunnel.uuid = uuid
      }
      tunnel.onstatechange = this.onTunnelStateChanged
      this.client = client
      this.tunnel = tunnel
      this.display = this.client.getDisplay()
      this.display.onresize = this.displayResize
      display.appendChild(client.getDisplay().getElement())
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      client.connect(connectionParams)

      window.onunload = function() {
        client.disconnect()
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