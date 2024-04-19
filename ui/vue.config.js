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
      }
    }
  },
  chainWebpack: config => {
    config
      .plugin('html')
      .tap(args => {
        args[0].title = '远程监查系统'
        return args
      })
  }
}
