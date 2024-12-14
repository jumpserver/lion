module.exports = {
  publicPath: '/lion/',
  outputDir: 'dist',
  assetsDir: 'assets',
  devServer: {
    port: 9529,
    proxy: {
      '^/lion/ws': {
        target: 'http://127.0.0.1:8081/',
        ws: true,
        changeOrigin: true
      },
      '^/lion/api': {
        target: 'http://127.0.0.1:8081/',
        ws: true,
        changeOrigin: true
      },
      '^/lion/token': {
        target: 'http://127.0.0.1:8081/',
        ws: true,
        changeOrigin: true
      },
      '^/api/v1': {
        target: 'http://127.0.0.1:8080/',
        ws: true,
        changeOrigin: true
      }
    }
  },
  chainWebpack(config) {
    config.resolve.alias.set('vue', '@vue/compat')
  }
}
