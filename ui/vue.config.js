module.exports = {
  publicPath: '/lion/',
  outputDir: 'lion',
  assetsDir: 'assets',
  devServer: {
    port: 9529,
    proxy: {
      '^/lion': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true
      }
    }
  }
}
