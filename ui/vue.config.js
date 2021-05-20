module.exports = {
  publicPath: '/lion/',
  outputDir: 'lion',
  assetsDir: 'assets',
  devServer: {
    port: 9529,
    proxy: {
      '^/lion': {
        target: 'http://192.168.1.47:8081',
        ws: true,
        changeOrigin: true
      }
    }
  },
  chainWebpack(config) {
  }
}
