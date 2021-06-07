module.exports = {
  publicPath: '/lion/',
  outputDir: 'lion',
  assetsDir: 'assets',
  devServer: {
    port: 9529,
    proxy: {
      '^/lion': {
        target: 'http://127.0.0.1:8081/',
        ws: true,
        changeOrigin: true
      }
    }
  },
  chainWebpack(config) {
  }
}
